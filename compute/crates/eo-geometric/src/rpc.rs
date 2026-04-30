//! Rational Polynomial Coefficient sensor model.
//!
//! The RPC model maps a normalised ground point `(P, L, H)` (latitude,
//! longitude, height) to a normalised image coordinate `(line, sample)`:
//!
//! ```text
//! line_n = NumL(P, L, H) / DenL(P, L, H)
//! samp_n = NumS(P, L, H) / DenS(P, L, H)
//! ```
//!
//! where each polynomial is a 20-term cubic in `(P, L, H)` with the canonical
//! term ordering specified by the OGC RPC sensor model. Normalisation:
//!
//! ```text
//! P = (lat   - lat_off)    / lat_scale
//! L = (lon   - lon_off)    / lon_scale
//! H = (height - height_off) / height_scale
//! line   = line_n * line_scale + line_off
//! sample = samp_n * samp_scale + samp_off
//! ```
//!
//! The 20 polynomial terms are ordered `[1, L, P, H, LP, LH, PH, L², P², H²,
//! LPH, L³, LP², LH², L²P, P³, PH², L²H, P²H, H³]`.

use crate::GeometricError;

/// Rational Polynomial Coefficient sensor model.
#[derive(Debug, Clone)]
pub struct Rpc {
    /// Image-space line offset.
    pub line_off: f64,
    /// Image-space sample offset.
    pub sample_off: f64,
    /// Geographic latitude offset (degrees).
    pub lat_off: f64,
    /// Geographic longitude offset (degrees).
    pub lon_off: f64,
    /// Height offset (metres).
    pub height_off: f64,
    /// Image-space line scale.
    pub line_scale: f64,
    /// Image-space sample scale.
    pub sample_scale: f64,
    /// Geographic latitude scale (degrees).
    pub lat_scale: f64,
    /// Geographic longitude scale (degrees).
    pub lon_scale: f64,
    /// Height scale (metres).
    pub height_scale: f64,
    /// 20 line-numerator coefficients.
    pub line_num: [f64; 20],
    /// 20 line-denominator coefficients.
    pub line_den: [f64; 20],
    /// 20 sample-numerator coefficients.
    pub samp_num: [f64; 20],
    /// 20 sample-denominator coefficients.
    pub samp_den: [f64; 20],
}

#[inline]
fn poly20(coef: &[f64; 20], p: f64, l: f64, h: f64) -> f64 {
    let l2 = l * l;
    let p2 = p * p;
    let h2 = h * h;
    coef[0]
        + coef[1] * l
        + coef[2] * p
        + coef[3] * h
        + coef[4] * l * p
        + coef[5] * l * h
        + coef[6] * p * h
        + coef[7] * l2
        + coef[8] * p2
        + coef[9] * h2
        + coef[10] * l * p * h
        + coef[11] * l2 * l
        + coef[12] * l * p2
        + coef[13] * l * h2
        + coef[14] * l2 * p
        + coef[15] * p2 * p
        + coef[16] * p * h2
        + coef[17] * l2 * h
        + coef[18] * p2 * h
        + coef[19] * h2 * h
}

impl Rpc {
    /// Forward projection: ground `(lat, lon, height)` → image `(line, sample)`.
    #[must_use]
    pub fn forward(&self, lat: f64, lon: f64, height: f64) -> (f64, f64) {
        let p = (lat - self.lat_off) / self.lat_scale;
        let l = (lon - self.lon_off) / self.lon_scale;
        let h = (height - self.height_off) / self.height_scale;
        let line_n = poly20(&self.line_num, p, l, h) / poly20(&self.line_den, p, l, h);
        let samp_n = poly20(&self.samp_num, p, l, h) / poly20(&self.samp_den, p, l, h);
        let line = line_n * self.line_scale + self.line_off;
        let sample = samp_n * self.sample_scale + self.sample_off;
        (line, sample)
    }

    /// Inverse projection: image `(line, sample)` at height `h_m` → ground
    /// `(lat, lon)`. Solved by Gauss-Newton on the 2-D residual; converges
    /// to sub-pixel accuracy in a few iterations for well-conditioned RPCs.
    ///
    /// # Errors
    /// Returns [`GeometricError::DidNotConverge`] if the iteration fails to
    /// reduce the residual below `tol_pixels` within `max_iters`.
    pub fn inverse(
        &self,
        line: f64,
        sample: f64,
        height: f64,
        tol_pixels: f64,
        max_iters: u32,
    ) -> Result<(f64, f64), GeometricError> {
        // Initial guess: model offsets.
        let mut lat = self.lat_off;
        let mut lon = self.lon_off;
        let eps_lat = self.lat_scale * 1e-6;
        let eps_lon = self.lon_scale * 1e-6;
        let mut last_residual = f64::INFINITY;

        for it in 0..max_iters {
            let (l0, s0) = self.forward(lat, lon, height);
            let r_line = line - l0;
            let r_samp = sample - s0;
            let residual = (r_line * r_line + r_samp * r_samp).sqrt();
            if residual < tol_pixels {
                return Ok((lat, lon));
            }
            // Numerical Jacobian.
            let (l_lat, s_lat) = self.forward(lat + eps_lat, lon, height);
            let (l_lon, s_lon) = self.forward(lat, lon + eps_lon, height);
            let dl_dlat = (l_lat - l0) / eps_lat;
            let ds_dlat = (s_lat - s0) / eps_lat;
            let dl_dlon = (l_lon - l0) / eps_lon;
            let ds_dlon = (s_lon - s0) / eps_lon;
            // Solve [J] * dx = r, J 2×2.
            let det = dl_dlat * ds_dlon - dl_dlon * ds_dlat;
            if det.abs() < 1e-18 {
                return Err(GeometricError::DidNotConverge { iters: it + 1, residual });
            }
            let dlat = (ds_dlon * r_line - dl_dlon * r_samp) / det;
            let dlon = (-ds_dlat * r_line + dl_dlat * r_samp) / det;
            lat += dlat;
            lon += dlon;
            last_residual = residual;
        }
        Err(GeometricError::DidNotConverge { iters: max_iters, residual: last_residual })
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    /// A synthetic identity-like RPC where the polynomials reduce to a linear
    /// mapping `(line, sample) = (lat, lon)` (with offsets/scales 0/1) so we
    /// can verify the algebra without a real product RPC file.
    fn identity_rpc() -> Rpc {
        let mut num_l = [0.0; 20];
        let mut den_l = [0.0; 20];
        let mut num_s = [0.0; 20];
        let mut den_s = [0.0; 20];
        // line_n = P (coefficient on P, index 2)
        num_l[2] = 1.0;
        den_l[0] = 1.0;
        // samp_n = L (coefficient on L, index 1)
        num_s[1] = 1.0;
        den_s[0] = 1.0;
        Rpc {
            line_off: 0.0,
            sample_off: 0.0,
            lat_off: 0.0,
            lon_off: 0.0,
            height_off: 0.0,
            line_scale: 1.0,
            sample_scale: 1.0,
            lat_scale: 1.0,
            lon_scale: 1.0,
            height_scale: 1.0,
            line_num: num_l,
            line_den: den_l,
            samp_num: num_s,
            samp_den: den_s,
        }
    }

    #[test]
    fn forward_identity() {
        let rpc = identity_rpc();
        let (line, samp) = rpc.forward(10.0, 20.0, 0.0);
        assert_abs_diff_eq!(line, 10.0, epsilon = 1e-12);
        assert_abs_diff_eq!(samp, 20.0, epsilon = 1e-12);
    }

    #[test]
    fn inverse_identity_recovers_ground() {
        let rpc = identity_rpc();
        let (lat, lon) = rpc.inverse(10.0, 20.0, 0.0, 1e-9, 50).unwrap();
        assert_abs_diff_eq!(lat, 10.0, epsilon = 1e-6);
        assert_abs_diff_eq!(lon, 20.0, epsilon = 1e-6);
    }

    #[test]
    fn inverse_round_trips_at_off_origin_point() {
        let rpc = identity_rpc();
        let (line, samp) = rpc.forward(-3.7, 5.1, 0.0);
        let (lat, lon) = rpc.inverse(line, samp, 0.0, 1e-9, 50).unwrap();
        assert_abs_diff_eq!(lat, -3.7, epsilon = 1e-6);
        assert_abs_diff_eq!(lon, 5.1, epsilon = 1e-6);
    }
}

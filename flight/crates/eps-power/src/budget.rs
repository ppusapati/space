//! Power budget bookkeeping with priority-based load shedding.

use crate::EpsError;

/// Single electrical load on the bus.
#[derive(Debug, Clone)]
pub struct Load {
    /// Stable, human-readable identifier.
    pub name: String,
    /// Demanded power (W).
    pub demand: f64,
    /// Priority — *lower numeric value = higher priority*. Loads with
    /// priority `< shed_priority_threshold` are guaranteed power as long
    /// as the total fits within the available budget.
    pub priority: u16,
}

/// Result of [`PowerBudget::allocate`].
#[derive(Debug, Clone)]
pub struct Allocation {
    /// Per-load delivered power (W), in the same order as the input.
    pub delivered: Vec<f64>,
    /// Total delivered (sum of `delivered`).
    pub total_delivered: f64,
    /// Total demanded (sum of input demands).
    pub total_demanded: f64,
    /// Whether any load was shed.
    pub shedding_active: bool,
}

/// Power budget allocator.
#[derive(Debug, Clone)]
pub struct PowerBudget {
    /// Available bus power (W). Negative values represent net battery
    /// drain — callers should monitor the battery service in that case.
    pub available_w: f64,
}

impl PowerBudget {
    /// Construct from the currently available bus power.
    ///
    /// # Errors
    /// [`EpsError::OutOfRange`] if `available_w` is non-finite.
    pub fn new(available_w: f64) -> Result<Self, EpsError> {
        if !available_w.is_finite() {
            return Err(EpsError::OutOfRange {
                name: "available_w",
                value: available_w,
                range: "finite",
            });
        }
        Ok(Self { available_w })
    }

    /// Allocate power across `loads` using priority-based shedding.
    /// Highest-priority loads (lowest `priority` value) are served first;
    /// when the running sum exceeds `available_w`, subsequent loads are
    /// shed (allocated 0 W) in priority order.
    ///
    /// # Errors
    /// [`EpsError::OutOfRange`] if any demand is negative or non-finite.
    pub fn allocate(&self, loads: &[Load]) -> Result<Allocation, EpsError> {
        for l in loads {
            if !(l.demand.is_finite() && l.demand >= 0.0) {
                return Err(EpsError::OutOfRange {
                    name: "load.demand",
                    value: l.demand,
                    range: "[0, +inf)",
                });
            }
        }
        // Sort indices by priority ascending; ties by original order.
        let mut order: Vec<usize> = (0..loads.len()).collect();
        order.sort_by_key(|&i| (loads[i].priority, i));
        let mut delivered = vec![0.0_f64; loads.len()];
        let mut sum = 0.0_f64;
        for &i in &order {
            let demand = loads[i].demand;
            if sum + demand <= self.available_w {
                delivered[i] = demand;
                sum += demand;
            } else {
                // Partial allocation to the marginal load if we have headroom.
                let headroom = (self.available_w - sum).max(0.0);
                if headroom > 0.0 {
                    delivered[i] = headroom;
                    sum += headroom;
                }
            }
        }
        let total_demanded: f64 = loads.iter().map(|l| l.demand).sum();
        let shedding_active = sum < total_demanded - 1e-12;
        Ok(Allocation { delivered, total_delivered: sum, total_demanded, shedding_active })
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    #[test]
    fn no_shedding_when_supply_exceeds_demand() {
        let b = PowerBudget::new(50.0).unwrap();
        let loads = vec![
            Load { name: "OBC".into(), demand: 5.0, priority: 1 },
            Load { name: "PAYLOAD".into(), demand: 20.0, priority: 5 },
        ];
        let a = b.allocate(&loads).unwrap();
        assert!(!a.shedding_active);
        assert_abs_diff_eq!(a.total_delivered, 25.0);
    }

    #[test]
    fn sheds_lowest_priority_first() {
        let b = PowerBudget::new(20.0).unwrap();
        let loads = vec![
            Load { name: "OBC".into(), demand: 5.0, priority: 1 },
            Load { name: "ADCS".into(), demand: 8.0, priority: 2 },
            Load { name: "PAYLOAD".into(), demand: 20.0, priority: 5 },
        ];
        let a = b.allocate(&loads).unwrap();
        assert!(a.shedding_active);
        assert_abs_diff_eq!(a.delivered[0], 5.0);
        assert_abs_diff_eq!(a.delivered[1], 8.0);
        // 7 W left → payload only gets partial.
        assert_abs_diff_eq!(a.delivered[2], 7.0);
        assert_abs_diff_eq!(a.total_delivered, 20.0);
    }

    #[test]
    fn rejects_negative_demand() {
        let b = PowerBudget::new(50.0).unwrap();
        let loads = vec![Load { name: "Bad".into(), demand: -1.0, priority: 1 }];
        assert!(b.allocate(&loads).is_err());
    }
}

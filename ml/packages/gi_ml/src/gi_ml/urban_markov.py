"""Urban-growth Markov-chain land-cover projection."""
from __future__ import annotations

import numpy as np


def project_landcover(
    landcover: np.ndarray, transition_matrix: np.ndarray, steps: int
) -> np.ndarray:
    """Project a categorical land-cover map forward by ``steps`` using a
    per-pixel Markov-chain transition.

    Args:
        landcover: ``(H, W)`` integer array with values in `[0, K)`.
        transition_matrix: ``(K, K)`` row-stochastic transition matrix
            where ``T[i, j]`` is the probability that a pixel currently
            of class `i` becomes class `j` after one step.
        steps: number of time steps to project.

    Returns the projected land-cover map (deterministic argmax over the
    accumulated class probability per pixel).

    Raises:
        ValueError: for malformed inputs.
    """
    lc = np.asarray(landcover)
    if lc.ndim != 2:
        raise ValueError("landcover must be 2-D")
    t = np.asarray(transition_matrix, dtype=np.float64)
    if t.ndim != 2 or t.shape[0] != t.shape[1]:
        raise ValueError("transition_matrix must be square")
    k = t.shape[0]
    if not (np.all(t >= -1e-12) and np.allclose(t.sum(axis=1), 1.0, atol=1e-6)):
        raise ValueError("transition_matrix must be row-stochastic")
    if int(lc.max()) >= k or int(lc.min()) < 0:
        raise ValueError("landcover values must lie in [0, K)")
    if steps <= 0:
        raise ValueError("steps must be positive")
    # Take T^steps in class space; project per pixel.
    t_power = np.linalg.matrix_power(t, steps)
    h, w = lc.shape
    out = np.zeros((h, w), dtype=lc.dtype)
    for i in range(h):
        row = lc[i]
        out[i] = np.argmax(t_power[row], axis=1)
    return out

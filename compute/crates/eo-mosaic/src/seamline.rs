//! Minimum-cost seamline through an overlap region.
//!
//! Given a per-pixel cost array (typically the absolute radiometric
//! difference between two registered tiles), find a path from any pixel
//! on one edge to any pixel on the opposite edge that minimises the total
//! cost. The output is a vector of `(row, col)` pixels — one per row when
//! the path runs top-to-bottom or one per column when left-to-right.

use std::cmp::Ordering;
use std::collections::BinaryHeap;

use ndarray::{Array2, ArrayView2};

/// Direction of seamline traversal.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Direction {
    /// Path runs from row 0 to the last row, with one pixel selected per row.
    TopToBottom,
    /// Path runs from column 0 to the last column, with one pixel selected per column.
    LeftToRight,
}

/// Minimum-cost path through `cost` in the given direction. Diagonal
/// neighbours are permitted; ties are broken by cost.
///
/// Returns the path in traversal order. Empty cost arrays return an empty
/// path.
#[must_use]
pub fn min_cost_path(cost: ArrayView2<'_, f32>, direction: Direction) -> Vec<(usize, usize)> {
    let (rows, cols) = cost.dim();
    if rows == 0 || cols == 0 {
        return Vec::new();
    }
    match direction {
        Direction::TopToBottom => path_top_to_bottom(cost, rows, cols),
        Direction::LeftToRight => path_left_to_right(cost, rows, cols),
    }
}

#[derive(Clone, Copy)]
struct Node {
    cost: f32,
    row: usize,
    col: usize,
}

impl Eq for Node {}

impl PartialEq for Node {
    fn eq(&self, other: &Self) -> bool {
        self.cost == other.cost && self.row == other.row && self.col == other.col
    }
}

impl Ord for Node {
    fn cmp(&self, other: &Self) -> Ordering {
        // Min-heap: invert cost ordering.
        other.cost.partial_cmp(&self.cost).unwrap_or(Ordering::Equal)
    }
}

impl PartialOrd for Node {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

fn path_top_to_bottom(
    cost: ArrayView2<'_, f32>,
    rows: usize,
    cols: usize,
) -> Vec<(usize, usize)> {
    // Dijkstra: source = virtual node above row 0 connected to every pixel
    // in row 0; sink = virtual node below row `rows-1` connected to every
    // pixel in the last row.
    let mut dist = Array2::<f32>::from_elem((rows, cols), f32::INFINITY);
    let mut prev = Array2::<i64>::from_elem((rows, cols), -1);
    let mut heap = BinaryHeap::new();
    for c in 0..cols {
        dist[(0, c)] = cost[(0, c)];
        heap.push(Node { cost: cost[(0, c)], row: 0, col: c });
    }
    while let Some(Node { cost: d, row, col }) = heap.pop() {
        if d > dist[(row, col)] {
            continue;
        }
        if row + 1 == rows {
            continue;
        }
        let nr = row + 1;
        let cmin = col.saturating_sub(1);
        let cmax = (col + 1).min(cols - 1);
        for nc in cmin..=cmax {
            let alt = d + cost[(nr, nc)];
            if alt < dist[(nr, nc)] {
                dist[(nr, nc)] = alt;
                prev[(nr, nc)] = (row * cols + col) as i64;
                heap.push(Node { cost: alt, row: nr, col: nc });
            }
        }
    }
    // Find best end pixel.
    let mut best_col = 0;
    let mut best = f32::INFINITY;
    for c in 0..cols {
        if dist[(rows - 1, c)] < best {
            best = dist[(rows - 1, c)];
            best_col = c;
        }
    }
    backtrack(prev.view(), rows - 1, best_col, cols)
}

fn path_left_to_right(
    cost: ArrayView2<'_, f32>,
    rows: usize,
    cols: usize,
) -> Vec<(usize, usize)> {
    let mut dist = Array2::<f32>::from_elem((rows, cols), f32::INFINITY);
    let mut prev = Array2::<i64>::from_elem((rows, cols), -1);
    let mut heap = BinaryHeap::new();
    for r in 0..rows {
        dist[(r, 0)] = cost[(r, 0)];
        heap.push(Node { cost: cost[(r, 0)], row: r, col: 0 });
    }
    while let Some(Node { cost: d, row, col }) = heap.pop() {
        if d > dist[(row, col)] {
            continue;
        }
        if col + 1 == cols {
            continue;
        }
        let nc = col + 1;
        let rmin = row.saturating_sub(1);
        let rmax = (row + 1).min(rows - 1);
        for nr in rmin..=rmax {
            let alt = d + cost[(nr, nc)];
            if alt < dist[(nr, nc)] {
                dist[(nr, nc)] = alt;
                prev[(nr, nc)] = (row * cols + col) as i64;
                heap.push(Node { cost: alt, row: nr, col: nc });
            }
        }
    }
    let mut best_row = 0;
    let mut best = f32::INFINITY;
    for r in 0..rows {
        if dist[(r, cols - 1)] < best {
            best = dist[(r, cols - 1)];
            best_row = r;
        }
    }
    backtrack(prev.view(), best_row, cols - 1, cols)
}

fn backtrack(
    prev: ArrayView2<'_, i64>,
    end_row: usize,
    end_col: usize,
    cols: usize,
) -> Vec<(usize, usize)> {
    let mut path = Vec::new();
    let mut cur = end_row * cols + end_col;
    loop {
        let row = cur / cols;
        let col = cur % cols;
        path.push((row, col));
        let p = prev[(row, col)];
        if p < 0 {
            break;
        }
        cur = p as usize;
    }
    path.reverse();
    path
}

#[cfg(test)]
mod tests {
    use ndarray::array;

    use super::*;

    #[test]
    fn ttb_chooses_cheapest_column() {
        // 3 columns, all rows: middle column cheap, edges expensive.
        let cost = array![[10.0_f32, 0.0, 10.0], [10.0, 0.0, 10.0], [10.0, 0.0, 10.0]];
        let path = min_cost_path(cost.view(), Direction::TopToBottom);
        assert_eq!(path.len(), 3);
        for &(_, c) in &path {
            assert_eq!(c, 1);
        }
    }

    #[test]
    fn ltr_chooses_cheapest_row() {
        let cost = array![[10.0_f32, 10.0, 10.0], [0.0, 0.0, 0.0], [10.0, 10.0, 10.0]];
        let path = min_cost_path(cost.view(), Direction::LeftToRight);
        assert_eq!(path.len(), 3);
        for &(r, _) in &path {
            assert_eq!(r, 1);
        }
    }

    #[test]
    fn empty_cost_yields_empty_path() {
        let cost = ndarray::Array2::<f32>::zeros((0, 0));
        assert!(min_cost_path(cost.view(), Direction::TopToBottom).is_empty());
    }
}

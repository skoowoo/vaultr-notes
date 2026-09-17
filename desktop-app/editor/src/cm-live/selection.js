// Inclusive: boundary caret counts — reveal raw markdown at the edge.
export function selectionTouchesRange(state, from, to) {
  for (const range of state.selection.ranges) {
    if (range.from <= to && range.to >= from) return true;
  }
  return false;
}

export function selectionTouchesLine(state, line) {
  return selectionTouchesRange(state, line.from, line.to);
}

// Strict (boundary-exclusive) for atomic widgets — inclusive would un-atomize
// on the same caret atomicRanges protects, breaking one-keystroke delete/skip.
export function selectionInsideRange(state, from, to) {
  for (const range of state.selection.ranges) {
    if (range.from > from && range.to < to) return true;
  }
  return false;
}

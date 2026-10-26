<?php
// Owned differential fixture for the actual-int64 binary subtraction subset.
// Overflow, generic unary negation, and decrement intentionally remain outside
// this fixture's successful domain.
function forward_integer($value) {
    return $value;
}

function offset_or_remaining($offset = null) {
    return $offset ?? strlen('abcd') - 1;
}

class OffsetBox {
    public $offset;
}

$box = new OffsetBox();
$box->offset = 2;

// Keep the oracle as raw stdout: the generated binary must preserve both the
// numeric values and PHP 8's lower concatenation precedence.
echo 'left=' . (20 - 7 - 3) . "\n";
echo 'mixed=' . (20 - 7 + 3) . "\n";
echo 'concat=' . 'x' . 10 - 3 . "\n";
echo 'negative=' . (5 - -3) . "\n";
echo 'zero=' . (0 - 0) . "\n";
echo 'forward=' . (forward_integer(9) - 4) . "\n";
echo 'coalesce=' . offset_or_remaining(0) . ':' . offset_or_remaining() . "\n";
echo 'property=' . (10 - $box->offset) . "\n";

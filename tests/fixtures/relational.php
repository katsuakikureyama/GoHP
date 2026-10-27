<?php
// Owned differential fixture for the intentionally narrow relational lane:
// < and > accept actual int64 and null values only. It exercises PHP's
// null/int bool conversion without admitting general loose comparison.

function relation_cell($left, $right) {
    return ($left < $right ? 'L' : '_')
        . ($left > $right ? 'G' : '_')
        . ($left === $right ? 'E' : '_');
}

function observed_once($value) {
    echo '<call>';
    return $value;
}

// The nested foreach drives every pair through the same untyped function.
// null compares as false: it equals 0, is below each nonzero integer
// (including a negative integer), and no other loose-comparison values are
// in this fixture's contract.
$values = [null, -2, 0, 2, 9223372036854775807];
$matrix = '';
foreach ($values as $left) {
    $row = '';
    foreach ($values as $right) {
        $row = $row . relation_cell($left, $right);
    }
    $matrix = $matrix . $row . ';';
}
echo 'matrix=' . $matrix . "\n";

// The minimum is spelled as the supported literal-plus-subtraction form;
// -9223372036854775808 is not a standalone source integer literal here.
$minimum = -9223372036854775807 - 1;
$maximum = 9223372036854775807;
echo 'bounds='
    . relation_cell($minimum, $maximum)
    . ':' . relation_cell($maximum, $minimum)
    . ':' . relation_cell($minimum, null)
    . ':' . relation_cell(null, $minimum)
    . "\n";

// Addition binds before relation, relation before strict identity, then
// &&, || and ternary. The parenthesized form separately makes the intended
// grouping explicit.
$precedence = 1 + 1 < 3 === true && 0 > null === false || false ? 'yes' : 'no';
$parenthesized = (($minimum < $maximum) && (($maximum > $minimum) === true)) ? 'yes' : 'no';
echo 'precedence=' . $precedence . ':' . $parenthesized . "\n";

// count() provides an actual int64 bound and the array supplies each
// int64/null value to the same untyped function.
$index = 0;
$count_trace = '';
while ($index < count($values)) {
    $count_trace = $count_trace . relation_cell($values[$index], $maximum);
    $index = $index + 1;
}
echo 'count=' . $index . ':' . $count_trace . "\n";

// This resembles the parser's offset/length state while keeping the class
// and properties locally declared and generic to this fixture.
class DocumentLike {
    public $offset = null;
    public $limit = null;
}

$document = new DocumentLike();
$document->offset = $minimum;
$document->limit = $maximum;
$document_state = ($document->offset < $document->limit && $document->limit > null) ? 'open' : 'closed';
echo 'property=' . $document_state . "\n";

// Both operands must run once. Their markers are deliberately identical, so
// this asserts count rather than PHP's unspecified operand evaluation order.
$single_evaluation = observed_once(-2) < observed_once(2);
echo 'single=' . ($single_evaluation ? 'yes' : 'no') . "\n";

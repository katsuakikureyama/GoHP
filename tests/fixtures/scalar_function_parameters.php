<?php
// Owned differential fixture for exact scalar parameter admission. PHP weak
// coercion, scalar return declarations, references, unions, and variadics are
// deliberately outside this bounded case.

function scalar_trace(InT $number, sTrInG $text, bOoL $flag, $tail = null) {
    $before = $number . ':' . $text . ':' . ($flag ? 'T' : 'F');
    $number = 'local-number';
    $text = false;
    $flag = 0;
    return $before . ':' . $number . ':' . ($text === false ? 'false' : 'wrong') . ':'
        . ($flag === 0 ? 'zero' : 'wrong') . ':' . ($tail === null ? 'null' : 'wrong');
}

function forward_scalar($number, $text, $flag, $tail = null) {
    return scalar_trace($number, $text, $flag, $tail);
}

function count_to(int $start, int $limit, string $tag, bool $early) {
    $current = $start;
    while ($current !== $limit) {
        if ($early) {
            return $tag . ':early:' . $current;
        }
        $current = $current + 1;
    }
    return $tag . ':done:' . $current;
}

function copy_with_scalars(array $items, int $amount, string $name, bool $enabled, $tail = null) {
    $items['local'] = $name . ':' . $amount . ':' . ($enabled ? 'T' : 'F');
    if ($tail === null) {
        $items[] = 'tail';
    }
    return $items;
}

function double_scalar(int $value) {
    return $value + $value;
}

$number = 9223372036854775807;
$text = '雪';
$flag = false;
$trace = scalar_trace($number, $text, $flag);
$forward = forward_scalar(-7, '', false, null);
$loop = count_to(-2, 0, "A\tB", false);
$early = count_to(7, 9, 'stop', true);
$copy_source = ['base' => 'kept'];
$copy = copy_with_scalars($copy_source, 3, 'copy', true);
$mapped = array_map('double_scalar', [0, 2, -3]);

echo 'trace=' . $trace . "\n";
echo 'caller=' . $number . ':' . $text . ':' . ($flag ? 'T' : 'F') . "\n";
echo 'forward=' . $forward . "\n";
echo 'flow=' . $loop . ':' . $early . "\n";
echo 'repeat=' . scalar_trace(0, '', true) . ':' . scalar_trace(-1, "A\tB", false) . "\n";
echo 'array=' . $copy_source['base'] . ':' . (isset($copy_source['local']) ? 'changed' : 'clean') . ':'
    . $copy['local'] . ':' . $copy[0] . "\n";
echo 'mapped=' . $mapped[0] . ':' . $mapped[1] . ':' . $mapped[2] . "\n";

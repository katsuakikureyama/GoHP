<?php
// Purpose-built GoHP coverage for a small, WordPress-like validation style.
// It is original fixture code, not a copied WordPress implementation.
function truth_word($value) {
    if ((bool) $value) {
        return 'T';
    }
    return 'F';
}

function kind_word($value) {
    return is_bool($value) . ':' . is_string($value);
}

function lower_guard($value) {
    if (is_string($value) && strtolower($value) === 'word') {
        return 'word';
    }
    return 'other';
}

function mark_once($value) {
    echo $value;
    return true;
}

function short_word($value) {
    if ($value && mark_once('!')) {
        return 'yes';
    }
    return 'no';
}

function early_word($value) {
    if (is_bool($value)) {
        if ($value === true) {
            return 'true';
        }
        return 'false';
    }
    return 'other';
}

function retained_word($value) {
    $state = 'cold';
    if ((bool) $value) {
        $state = 'warm';
    }
    return $state;
}

function strict_word($value) {
    if ($value === '1') {
        return 'string';
    }
    return 'other';
}

echo 'bool=' . truth_word(false) . truth_word(null) . truth_word(0) . truth_word('0') . truth_word('') . truth_word('false') . truth_word('00') . truth_word([1]) . "\n";
echo 'pred=' . kind_word(false) . ',' . kind_word('text') . ',' . kind_word([]) . "\n";
echo 'lower=' . lower_guard(7) . ',' . lower_guard([]) . ',' . lower_guard('WORD') . ',' . lower_guard('word') . "\n";
echo 'short=' . short_word(false) . ':' . short_word(true) . "\n";
echo 'early=' . early_word(true) . ',' . early_word(false) . ',' . early_word('true') . "\n";
echo 'retained=' . retained_word(false) . ',' . retained_word(true) . "\n";
echo 'strict=' . strict_word(1) . ',' . strict_word('1') . "\n";
echo 'precedence=' . (1 . 2 + 3 === 15 && true) . "\n";
echo 'empty=' . truth_word([]) . ',parenthesized=' . ((true === false) === false) . ',cast=' . ((bool) '0' === false) . "\n";

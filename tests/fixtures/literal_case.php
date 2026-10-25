<?php
function bool_word($value) {
    return $value ? 'T' : 'F';
}

function mixed_return($value) {
    return $value ? TrUe : nUlL;
}

function condition_word() {
    if (TrUe && !fAlSe) {
        return 'condition';
    }
    return 'other';
}

function true_value() {
    return 'function-name';
}

class LiteralDefaults {
    public $yes = tRuE;
    public $YES = FaLsE;
    public $no = fAlSe;
    public $none = NuLl;
}

$entries = ['null' => nUlL, 'false' => FaLsE, 'true' => TrUe];
$defaults = new LiteralDefaults();
$TRUE = 'upper-variable';
$true = 'lower-variable';

echo 'return=' . bool_word(mixed_return(TrUe)) . ':' . bool_word(mixed_return(FaLsE)) . "\n";
echo 'defaults=' . bool_word($defaults->yes) . ':' . bool_word($defaults->no) . ':' . bool_word($defaults->none) . "\n";
echo 'isset=' . bool_word(isset($entries['null'])) . ':' . bool_word(isset($entries['false'])) . ':' . bool_word(isset($entries['true'])) . "\n";
echo 'strict=' . bool_word(in_array(TrUe, $entries, tRuE)) . ':' . bool_word(in_array(FaLsE, $entries, TRUE)) . "\n";
echo 'logic=' . bool_word(TrUe && !fAlSe) . ':' . bool_word(FaLsE || NuLl) . ':' . (TrUe ? 'then' : 'else') . "\n";
echo 'condition=' . condition_word() . ':' . (FaLsE ? 'then' : 'else') . "\n";
echo 'names=' . $TRUE . ':' . $true . ':' . 'TRUE' . ':' . true_value() . "\n";
echo 'types=' . bool_word(mixed_return(FALSE) === NULL) . ':' . bool_word(mixed_return(TRUE) === TRUE) . ':' . bool_word($defaults->no === FALSE) . ':' . bool_word($defaults->none === NULL) . ':' . bool_word('TRUE' !== TRUE) . "\n";
echo 'property-case=' . bool_word($defaults->yes === TRUE) . ':' . bool_word($defaults->YES === FALSE) . "\n";

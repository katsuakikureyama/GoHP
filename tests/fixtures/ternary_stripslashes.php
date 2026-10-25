<?php
function marker($value) {
    echo '<' . $value . '>';
    return $value;
}

function stripslashes_from_strings_only($value) {
    return is_string($value) ? StRiPsLaShEs($value) : $value;
}

echo '[' . (true && false ? 'yes' : 'no') . "]\n";
echo '[' . (1 === 1 ? 'equal' : 'not-equal') . "]\n";
echo '[' . (1 + 2 ? 'sum' : 'zero') . "]\n";
echo '[' . ('' . 'x' ? 'concat' : 'empty') . "]\n";
echo '[' . (true ? false ? 'middle-yes' : 'middle-no' : 'outer-no') . "]\n";
echo '[' . ((false ? 'left-yes' : 'left-no') ? 'outer-yes' : 'outer-no') . "]\n";
echo '[' . (false ? 'right-no' : (true ? 'right-yes' : 'right-inner-no')) . "]\n";
echo '[' . (marker('condition') ? marker('yes') : marker('never')) . "]\n";
echo '[' . (marker('0') ? marker('never-again') : marker('no')) . "]\n";
echo '[' . stripslashes_from_strings_only('O\\\'Brien') . "]\n";
echo '[' . stripslashes_from_strings_only('a\\\\b') . "]\n";
echo '[' . stripslashes_from_strings_only('a\\qb') . "]\n";
echo '[' . stripslashes_from_strings_only('A\\0B') . "]\n";
echo '[' . stripslashes_from_strings_only('tail\\') . "]\n";
echo '[' . stripslashes_from_strings_only(42) . "]\n";
echo '[' . stripslashes_from_strings_only(null) . "]\n";

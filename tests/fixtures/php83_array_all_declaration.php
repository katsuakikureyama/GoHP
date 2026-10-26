<?php
// This is a user declaration, not an implementation of PHP 8.4's array_all
// builtin. PHP 8.3 permits the name; PHP 8.4 reserves it.

function ArRaY_AlL($value) {
    if (gettype($value) === 'string') {
        return 'user:' . $value;
    }
    return false;
}

$text = aRrAy_AlL('owned');
$other = ARRAY_ALL(7);

echo 'value=' . $text . ':' . gettype($text) . "\n";
echo 'false=' . ($other === false ? 'true' : 'wrong') . ':' . gettype($other) . "\n";

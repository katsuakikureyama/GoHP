<?php
// This is a user declaration, not an implementation of PHP 8.4's array_any
// builtin. PHP 8.3 permits the name; PHP 8.4 reserves it.

function ArRaY_AnY($value) {
    if (gettype($value) === 'string') {
        return 'user:' . $value;
    }
    return false;
}

$text = aRrAy_AnY('owned');
$other = ARRAY_ANY(7);

echo 'value=' . $text . ':' . gettype($text) . "\n";
echo 'false=' . ($other === false ? 'true' : 'wrong') . ':' . gettype($other) . "\n";

<?php
function string_length($value) {
    return strlen($value);
}
function string_value() {
    return 'returned';
}
$words = ['item' => 'lookup'];
echo strlen("") . "\n";
echo strlen("ASCII") . "\n";
echo strlen(" \t\n\r") . "\n";
echo strlen('a\nb') . "\n";
echo strlen("こんにちは") . "\n";
echo strlen("🙂") . "\n";
echo strlen('go' . 'hp') . "\n";
echo STRLEN('Case') . "\n";
echo string_length('return') . "\n";
echo strlen(string_value()) . "\n";
echo strlen($words['item']) . "\n";
echo strlen('abc') + 2 . "\n";

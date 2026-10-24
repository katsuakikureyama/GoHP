<?php
function replace_text($search, $replacement, $subject) {
    return str_replace($search, $replacement, $subject);
}

echo '[' . replace_text('cat', 'dog', 'cat cat') . "]\n";
echo '[' . replace_text('', 'X', 'abc') . "]\n";
echo '[' . replace_text('aa', 'X', 'aaaa') . "]\n";
echo '[' . replace_text('a', 'aa', 'a') . "]\n";
echo '[' . replace_text('-', '', 'a-b-c') . "]\n";
echo '[' . StR_RePlAcE('A', 'x', 'aA') . "]\n";
echo '[' . replace_text('猫', '犬', '猫と猫') . "]\n";
echo '[' . replace_text('x', '$1', 'xx') . "]\n";

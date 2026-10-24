<?php
function zeroise($number, $threshold) {
    return sprintf('%0' . $threshold . 's', $number);
}

echo '[' . zeroise(42, 5) . "]\n";
// Negative integer literals are outside the parser subset; Gossip's typed
// observer supplies negative int64 values directly to the generated function.
echo '[' . zeroise('-42', 5) . "]\n";
echo '[' . zeroise('007', 5) . "]\n";
echo '[' . zeroise('', 3) . "]\n";
echo '[' . zeroise(0, 0) . "]\n";
echo '[' . zeroise('ready', 3) . "]\n";
echo '[' . zeroise('é', 3) . "]\n";

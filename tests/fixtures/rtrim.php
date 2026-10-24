<?php
function trim_slashes($value) {
    return rtrim($value, '/');
}

echo '[' . rtrim('', '/') . "]\n";
echo '[' . rtrim('plain', '/') . "]\n";
echo '[' . rtrim('path///', '/') . "]\n";
echo '[' . rtrim('path\\\\', '\\\\') . "]\n";
echo '[' . rtrim('xyxy', 'xy') . "]\n";
echo '[' . rtrim('stay', '') . "]\n";
echo '[' . trim_slashes('nested///') . "]\n";
echo '[' . RtRiM('mixed...', '.') . "]\n";
echo '[' . rtrim('é', '©') . "]\n";

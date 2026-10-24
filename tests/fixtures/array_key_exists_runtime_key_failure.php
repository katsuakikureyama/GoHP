<?php
function has_key($key, $value) {
    return array_key_exists($key, $value);
}
echo has_key(true, ['x' => 1]);

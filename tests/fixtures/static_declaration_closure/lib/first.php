<?php
function first_touch(SharedValue $item) {
    $item->value = $item->value . ':first';
    return $item->value;
}

require_once __DIR__ . '/nested/shared.php';

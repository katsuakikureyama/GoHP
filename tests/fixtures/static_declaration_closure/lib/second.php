<?php
function second_read(SharedValue $item) {
    return $item->value;
}

require_once __DIR__ . '/nested/shared.php';

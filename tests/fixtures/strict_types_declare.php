<?php
declare(strict_types=1);

// Differential fixture: the strict_types declaration is a leading no-op for
// the exact-typed subset.
function label(string $name, int $count) {
    return $name . ":" . $count;
}

echo label("items", 3) . "\n";
echo "done\n";

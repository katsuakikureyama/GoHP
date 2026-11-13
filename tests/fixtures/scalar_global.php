<?php
// Differential fixture for scalar-only PHP global statement support.
// Tests cross-call writes, main/function sharing, missing import -> null,
// two names in one statement, local shadowing, and type coverage.

function f() {
    global $counter;
    $counter = $counter + 1;
}

function g() {
    global $counter;
    return $counter;
}

function writeGlobal() {
    global $shared;
    $shared = 'from-function';
}

function readGlobal() {
    global $shared;
    return $shared;
}

function twoNames() {
    global $a, $b;
    $a = 10;
    $b = 'hello';
}

function localShadow() {
    // No global declaration — $counter is a local variable here.
    $counter = 999;
    return $counter;
}

// Cross-call write in f() then read in g().
f();
f();
echo 'cross-call:' . g() . "\n";

// Main writes, function reads.
global $shared;
$shared = 'from-main';
echo 'main-write-func-read:' . readGlobal() . "\n";

// Function writes, main reads.
writeGlobal();
global $shared2;
$shared2 = $shared;
echo 'func-write-main-read:' . $shared2 . "\n";

// Missing global imported -> null.
global $missing;
echo 'missing-type:' . gettype($missing) . "\n";
echo 'missing-null:' . ($missing === null ? 'true' : 'false') . "\n";

// Two names in one statement.
twoNames();
global $a, $b;
echo 'two-names:' . $a . ':' . $b . "\n";

// Local shadowing in a function without global.
echo 'local-shadow:' . localShadow() . "\n";
echo 'counter-unchanged:' . g() . "\n";

// Top-level global -> null.
global $y;
echo 'top-level-null:' . gettype($y) . "\n";

// Type coverage: int, string, bool, null.
global $types;
$types = 42;
echo 'int-type:' . gettype($types) . ':' . $types . "\n";
$types = 'text';
echo 'string-type:' . gettype($types) . ':' . $types . "\n";
$types = true;
echo 'bool-type:' . gettype($types) . ':' . ($types ? 'true' : 'false') . "\n";
$types = null;
echo 'null-type:' . gettype($types) . "\n";

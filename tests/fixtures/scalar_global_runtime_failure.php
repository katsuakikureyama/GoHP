<?php
// Runtime failure fixture: assigning an array to a global-bound variable
// must fail explicitly via the scalar global guard.
function setArray() {
    global $arr;
    $arr = [1, 2, 3];
}
setArray();

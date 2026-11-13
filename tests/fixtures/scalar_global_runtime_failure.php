<?php
// Runtime failure fixture: a global-bound variable that holds an integer
// cannot be written as an array; the generated program must fail explicitly.
function setCount() {
    global $count;
    $count = 5;
}
function setElement() {
    global $count;
    $count[0] = 1;
}
setCount();
setElement();

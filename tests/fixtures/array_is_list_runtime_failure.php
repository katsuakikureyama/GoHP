<?php
function list_for($value) {
    return array_is_list($value);
}
echo list_for('not an array');

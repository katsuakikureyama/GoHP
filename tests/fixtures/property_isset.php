<?php
// Owned differential fixture for named-property isset(). Every property path
// begins with a defined bare local (or $this in a method). This keeps the
// corpus focused on PHP's quiet existence/null check rather than dynamic
// names, magic accessors, indexing, or effectful receiver expressions.

function bool_word($value) {
    if ($value) {
        return 'T';
    }
    return 'F';
}

function dynamic_child($value) {
    return isset($value->child);
}

function unknown_value($value) { return $value; }

class IssetLeaf {
    public $value = null;
    public $present = false;
}

class IssetProbe {
    public $implicit_null;
    public $explicit_null = null;
    public $false_value = false;
    public $zero_value = 0;
    public $empty_value = '';
    public $zero_text = '0';
    public $array_value = null;
    public $child = null;

    public function from_this() {
        return isset($this->child->value);
    }
}

$leaf = new IssetLeaf();
$leaf->value = 'seed';
$probe = new IssetProbe();
$probe->child = $leaf;
$probe->array_value = [];
$alias = $probe;
$alias->child->value = 'shared';

$null_child = new IssetProbe();
$scalar_child = new IssetProbe();
$scalar_child->child = 'scalar';
$missing_leaf = new IssetLeaf();
$missing_probe = new IssetProbe();
$missing_probe->child = $missing_leaf;

$null_value = null;
$false_value = false;
$items = ['present' => false, 'null' => null];
$present_key = 'present';
$poison = unknown_value('not-an-array');

// The third operand must not be read after the missing property makes the
// multi-address check false. The true case mixes bare, indexed and property
// addresses, all of which are non-null even when their values are falsy.
$multi_true = isset($false_value, $items[$present_key], $probe->false_value);
$multi_false = isset($false_value, $probe->missing, $poison['never-read']);
$logical = isset($probe->missing) && isset($poison->never_read);
$ternary = isset($probe->child->value) ? 'yes' : 'no';
$branch = 'unset';
if (isset($probe->child->value)) {
    $branch = 'then';
} else {
    $branch = 'else';
}

echo 'declared=' . bool_word(isset($probe->implicit_null)) . ':' . bool_word(isset($probe->explicit_null)) . ':' . bool_word(isset($probe->false_value)) . ':' . bool_word(isset($probe->zero_value)) . ':' . bool_word(isset($probe->empty_value)) . ':' . bool_word(isset($probe->zero_text)) . ':' . bool_word(isset($probe->array_value)) . "\n";
echo 'chains=' . bool_word(isset($probe->child->value)) . ':' . bool_word($probe->from_this()) . ':' . bool_word(isset($probe->missing)) . ':' . bool_word(isset($missing_probe->child->missing)) . ':' . bool_word(isset($null_child->child->value)) . ':' . bool_word(isset($scalar_child->child->value)) . "\n";
echo 'dynamic=' . bool_word(dynamic_child(null)) . ':' . bool_word(dynamic_child(false)) . ':' . bool_word(dynamic_child(0)) . ':' . bool_word(dynamic_child('text')) . ':' . bool_word(dynamic_child([])) . "\n";
echo 'shared=' . $probe->child->value . ':' . $alias->child->value . "\n";
echo 'multi=' . bool_word($multi_true) . ':' . bool_word($multi_false) . ':' . bool_word($logical) . ':' . $ternary . ':' . $branch . "\n";

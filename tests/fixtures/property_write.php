<?php
// Owned differential fixture: only declared public, untyped properties receive
// writes through known generated-object variables or $this in an untyped method.
class PropertyStore {
    public $supports = null;
    public $state = null;
    public $label = null;
    public $Label = 'upper';
    public $payload = null;
    public $callback = null;

    public function write_from_method($value) {
        $this->state = $value;
        return $this->state;
    }
}

class PropertyPayload {
    public $name = 'original';
}

function write_marker($value) {
    echo '<marker:' . $value . '>';
    return $value;
}

// This mirrors the planned Gossip shape: an unknown input parameter is stored
// on a fresh, locally known generated object, then read back through its getter.
function copy_to_new_supports($input) {
    $local = new PropertyStore();
    $local->supports = $input;
    return $local->supports;
}

$source = ['nested' => ['child' => 'source']];
$store = new PropertyStore();
$store->supports = $source;
$source['nested'] = ['child' => 'source-mutated'];
$stored_before_copy = $store->supports;
$getter_copy = $store->supports;
$getter_copy['nested'] = ['child' => 'getter-copy'];
$stored_after_copy = $store->supports;

$alias = $store;
$alias->supports = ['nested' => ['child' => 'alias']];
$shared_after_alias = $store->supports;

$method_result = $store->write_from_method('method');
$method_state = $store->state;
$store->state = null;
$state_null = $store->state;
$store->state = 7;
$state_int = $store->state;
$store->state = 'text';
$state_string = $store->state;
$store->state = ['value' => 'array'];
$state_array = $store->state;

$store->label = write_marker('once');
$marked = $store->label;
$store->Label = 'changed-upper';
$case_upper = $store->Label;
$case_lower = $store->label;

$payload = new PropertyPayload();
$store->payload = $payload;
$payload->name = 'changed-through-alias';
$payload_from_store = $store->payload;
$arrow = fn($value) => $value;
$store->callback = $arrow;
$callback_from_store = $store->callback;
$arrow_value = $callback_from_store('arrow-result');

$driver_input = ['nested' => ['child' => 'driver-input']];
$driver_result = copy_to_new_supports($driver_input);
$driver_result['nested'] = ['child' => 'driver-result-copy'];

echo 'copies:' . $stored_before_copy['nested']['child'] . ':' . $source['nested']['child'] . ':' . $getter_copy['nested']['child'] . ':' . $stored_after_copy['nested']['child'] . "\n";
echo 'alias:' . $shared_after_alias['nested']['child'] . "\n";
echo 'method:' . $method_result . ':' . $method_state . "\n";
echo 'transitions:' . (is_null($state_null) ? 'null' : 'wrong') . ':' . $state_int . ':' . $state_string . ':' . $state_array['value'] . "\n";
echo 'case:' . $case_upper . ':' . $case_lower . ':' . $marked . "\n";
echo 'objects:' . (is_object($payload_from_store) ? 'object' : 'wrong') . ':' . $payload_from_store->name . ':' . (is_object($callback_from_store) ? 'object' : 'wrong') . ':' . $arrow_value . "\n";
echo 'driver:' . $driver_input['nested']['child'] . ':' . $driver_result['nested']['child'] . "\n";

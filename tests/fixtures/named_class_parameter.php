<?php
// Differential coverage for non-null, bare generated-class parameters. The
// first annotation deliberately precedes Node and changes case, so admission
// must collect declarations and canonicalize the class name before validating
// either body or call site.
function global_trace(nOdE $node, string $label, $tail = null) {
    $nested = ['inner' => ['label' => $label], 'node' => $node];
    $copy = $nested;
    $copy['inner'] = ['label' => 'local-copy'];

    // Arrays copy by value while their object elements retain identity. The
    // parameter may subsequently be rebound without replacing the caller.
    $node->name = $label . '-mutated';
    $embedded = $copy['node'];
    $node = new Node('global-local');
    return $copy['inner']['label'] . ':' . $nested['inner']['label'] . ':' . $embedded->name . ':' . $node->name;
}

class Node {
    public $name = null;

    public function __construct($name) {
        $this->name = $name;
    }
}

class Holder {
    public $node = null;
    public $label = null;
    public $tail = null;

    // Constructors may combine the admitted named class with existing exact
    // scalar entries and an untyped optional null suffix.
    public function __CoNsTrUcT(NoDe $node, string $label, $tail = null) {
        $this->node = $node;
        $this->label = $label;
        $this->tail = $tail;
    }

    public function touch(nOdE $node, $label, $tail = null) {
        $node->name = $label;
        $node = new Node('method-local');
        return $this->node->name . ':' . $node->name . ':' . $this->label;
    }

    // This keeps the typed-method entry guard on a nested Execution path.
    public function token(NoDe $node, $source) {
        $node->name = strtok($source, ',');
        return $node->name;
    }
}

$shared = new NODE('seed');
$global = global_trace($shared, 'global');
$holder = new holder($shared, 'frame');
$method = $holder->touch($shared, 'method');
$token = $holder->token($shared, 'first,second');
$after = strtok(',');

echo 'global=' . $global . ':' . $shared->name . "\n";
echo 'method=' . $method . ':' . $holder->node->name . "\n";
echo 'tokens=' . $token . ':' . $after . ':' . $shared->name . "\n";

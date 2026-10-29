<?php
// Owned differential fixture for non-null declared class parameters. Class
// declarations intentionally follow their consumers: admission resolves the
// complete generated-class set rather than PHP source order. Nullable, union,
// inherited, and reference parameters remain outside this bounded contract.

function FoRwArD_tOuCh(iTeM $item, int $count, $tail = null) {
    $item->value = $item->value . ':F' . $count;
    return $item->value;
}

function mapped_touch(ITEM $item) {
    $item->value = $item->value . '!';
    return $item->value;
}

function token_touch(Item $item, int $count, $tail = null) {
    $item->value = $item->value . ':T' . $count;
    return strtok('first,next', ',');
}

function local_reassignment(item $item) {
    $item = new ITEM('local-only');
    return $item;
}

class Holder {
    public $item;
    public $tag;

    public function __construct(ITEM $item, string $tag, $tail = null) {
        echo '<ctor>';
        $this->item = $item;
        $this->tag = $tag;
        $item->value = $item->value . ':C-' . $tag;
    }

    public function touch(iTeM $item, $count, $tail = null) {
        $item->value = $item->value . ':M' . $count;
        return $item->value;
    }
}

class Item {
    public $value;

    public function __construct($value) {
        $this->value = $value;
    }
}

$item = new iTeM('seed');
$forward = fOrWaRd_ToUcH($item, 2);
$holder = new hOlDeR($item, 'holder');
$method = $holder->touch($item, 3);
$mapped = array_map('MAPPED_TOUCH', [$item]);
$token_first = TOKEN_TOUCH($item, 4);
$token_second = strtok(',');
$replacement = local_reassignment($item);
$held = $holder->item;

echo 'trace=' . $forward . ':' . $method . ':' . $mapped[0] . ':' . $token_first . ':' . $token_second . "\n";
echo 'objects=' . $held->value . ':' . $item->value . ':' . $replacement->value . ':' . $holder->tag . "\n";

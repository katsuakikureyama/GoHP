<?php
// Owned differential fixture for the bounded public-constructor scalar
// annotations. Every annotated argument below has its exact PHP 8.3 type;
// coercion and annotation failures intentionally stay outside this fixture.
function argument_marker($name, $value) {
    echo '<arg:' . $name . '>';
    return $value;
}

function bool_word($value) {
    return $value ? 'true' : 'false';
}

class TypedPacket {
    public $integer;
    public $string;
    public $flag;
    public $label;
    public $optional;
    public $initial;

    public function __CoNsTrUcT(InT $integer, sTrInG $string, bOoL $flag, $label, $optional = null) {
        // Declared public properties exist as null before the constructor body.
        $this->initial = bool_word(
            is_null($this->integer) && is_null($this->string) && is_null($this->flag)
        );
        $this->integer = $integer;
        $this->string = $string;
        $this->flag = $flag;
        $this->label = $label;
        $this->optional = $optional ?? 'default';

        // Constructor annotations check entry only; locals can change type.
        $integer = 'local-int';
        $string = 7;
        $flag = 'local-bool';
        echo '<body:' . $integer . ':' . $string . ':' . $flag . '>';
    }
}

class TypedTokenizer {
    public $label;
    public $enabled;
    public $token;

    public function __construct(STRING $label, BOOL $enabled) {
        $this->label = $label;
        $this->enabled = $enabled;
        $this->token = strtok(',');
    }
}

// Arguments, including false and the empty string, run before the constructor.
$first = new typedpacket(
    argument_marker('integer', 0),
    argument_marker('string', ''),
    argument_marker('flag', false),
    argument_marker('label', 'first')
);
$second = new TypedPacket(-7, '0', true, 'second', 'provided');
$large = new TYPEDPACKET(9223372036854775807, "雪\f", true, 'large');

$first_snapshot = (array) $first;
$first->label = 'first-mutated';

$token_first = strtok('caller,constructor,after', ',');
$tokenized = new TYPEDTOKENIZER('token', true);
$token_after = strtok(',');

echo "\nfirst=" . $first->integer . ':' . $first->string . ':' . bool_word($first->flag)
    . ':' . $first->label . ':' . $first->optional . ':' . $first->initial . "\n";
echo 'second=' . $second->integer . ':' . $second->string . ':' . bool_word($second->flag)
    . ':' . $second->label . ':' . $second->optional . ':' . $second->initial . "\n";
echo 'large=' . $large->integer . ':' . $large->string . ':' . bool_word($large->flag)
    . ':' . $large->label . ':' . $large->optional . ':' . $large->initial . "\n";
echo 'snapshot=' . $first_snapshot['integer'] . ':' . $first_snapshot['label'] . ':'
    . $first->label . ':' . $first_snapshot['optional'] . "\n";
echo 'shared=' . $token_first . ':' . $tokenized->label . ':' . bool_word($tokenized->enabled)
    . ':' . $tokenized->token . ':' . $token_after . "\n";

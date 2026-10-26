<?php
// Owned differential fixture for the deliberately narrow parser-output path.
// The parser stores only actual arrays in its public output property.  Each
// freeform datum is constructed by a one-statement factory method and then
// exposed through PHP's generated-public-property array cast.
class DataBlock {
    public $content;

    public function __construct($content) {
        $this->content = $content;
    }
}

class OwnParser {
    public $output;

    public function __construct($output = null) {
        $this->output = $output ?? [];
    }

    public function freeform($html, $start = null) {
        return new DataBlock(substr($html, $start ?? 0));
    }

    public function add_freeform($html, $start = null) {
        $this->output[] = (array) $this->freeform($html, $start);
    }

    public function probe($stop = null) {
        if ($stop) {
            return;
        }
    }

    public function next_token($separator = null) {
        return strtok($separator ?? ',');
    }

    public function reset_output() {
        $this->output = ['reset'];
        return ['rhs'];
    }

    public function append_after_reset() {
        $this->output[] = $this->reset_output();
    }
}

$seed = ['seed'];
$parser = new OwnParser($seed);
$seed[] = 'caller-mutated';
$parser->add_freeform('abcd', 1);
$parser->add_freeform('xy');

$first_token = strtok('caller,method,after', ',');
$second_token = $parser->next_token();
$third_token = $parser->next_token(null);

$reset = new OwnParser(['initial']);
$reset->append_after_reset();

echo 'copy=' . $parser->output[0] . ':' . count($parser->output) . ':' . count($seed) . "\n";
echo 'output=' . $parser->output[1]['content'] . ':' . $parser->output[2]['content'] . "\n";
echo 'probe=' . (is_null($parser->probe(true)) ? 'null' : 'wrong') . ':' . (is_null($parser->probe()) ? 'null' : 'wrong') . "\n";
echo 'tokens=' . $first_token . ':' . $second_token . ':' . $third_token . "\n";
echo 'reset=' . count($reset->output) . ':' . $reset->output[0] . ':' . $reset->output[1][0] . "\n";
echo 'substr=' . substr('abcd', 1) . ':' . substr('abcd', -2) . ':' . substr('abcd', 0, 0) . ':' . substr('abcd', 99) . ':' . substr('abcd', 1, 2) . ':' . substr('abcd', 1, -1) . ':' . substr('abcd', 1, null) . ':' . substr('é', 1, 1) . "\n";

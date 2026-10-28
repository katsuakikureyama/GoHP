package tests

import "testing"

func TestNestedPropertyRuntimeGuards(t *testing.T) {
	checkArrayRuntimeGuards(t, []arrayRuntimeCase{
		{
			"dynamic_root_scalar",
			"<?php function update($frame) { $frame->block->text = 'value'; } update('scalar');",
			"nested property requires object",
		},
		{
			"null_intermediate",
			"<?php class Block { public $text = null; } class Frame { public $block = null; } $frame = new Frame(); $frame->block->text = 'value';",
			"nested property requires object",
		},
		{
			"append_non_array",
			"<?php class Block { public $items = null; } class Frame { public $block = null; } $frame = new Frame(); $frame->block = new Block(); $frame->block->items[] = 'value';",
			"array append requires array",
		},
		{
			"concat_non_string",
			"<?php class Block { public $text = 7; } class Frame { public $block = null; } $frame = new Frame(); $frame->block = new Block(); $frame->block->text .= 'value';",
			"nested property concatenation requires strings",
		},
	})
}

<?php
// Owned differential fixture for the narrow json_decode($text, true) lane.
// Every successful numeric token is an int64. Fractions, exponents, and values
// outside int64 are deliberately rejected by this profile, not PHP-oracled here.

function decode_error_code($text) {
    $decoded = json_decode($text, true);
    return json_last_error();
}

function map_decode_value($text) {
    return json_decode($text, true);
}

class JsonErrorProbe {
    public function decode_error_code($text) {
        $decoded = json_decode($text, true);
        return json_last_error();
    }
}

// Root scalar observations retain their actual PHP value type rather than
// coercing through a serializer.
$root_int = json_decode('-9223372036854775808', true);
$root_string = json_decode('"root"', true);
$root_bool = json_decode('true', true);
$root_null = json_decode('null', true);
echo 'roots:' . $root_int . ':' . $root_string . ':' . ($root_bool ? 'true' : 'false') . ':' . (is_null($root_null) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";

// This is shaped like a block attribute payload: nested objects/lists, false,
// null, an empty object, and an empty list all travel through associative mode.
$attrs = json_decode('{"title":"\\u732b","items":[{"id":1,"enabled":true},null],"visible":false,"empty_object":{},"empty_list":[]}', true);
echo 'attrs:' . $attrs['title'] . ':' . $attrs['items'][0]['id'] . ':' . ($attrs['items'][0]['enabled'] ? 'true' : 'false') . ':' . (is_null($attrs['items'][1]) ? 'null' : 'wrong') . ':' . ($attrs['visible'] ? 'true' : 'false') . ':' . count($attrs['empty_object']) . ':' . count($attrs['empty_list']) . "\n";

// Object insertion order stays at the first spelling of a duplicate key while
// its value is replaced. Canonical integer-string keys use normal PHP keys;
// leading-zero and out-of-range spellings remain strings.
$keys_value = json_decode('{"b":1,"0":"zero","-1":"negative","08":"leading","9223372036854775807":"maximum","9223372036854775808":"overflow-key","-9223372036854775808":"minimum","-9223372036854775809":"underflow-key","b":3}', true);
$keys = array_keys($keys_value);
echo 'keys:' . count($keys_value) . ':' . $keys[0] . ':' . $keys[1] . ':' . $keys[2] . ':' . $keys[3] . ':' . $keys[4] . ':' . $keys[5] . ':' . $keys[6] . ':' . $keys[7] . ':' . $keys_value['b'] . ':' . $keys_value[0] . ':' . $keys_value[-1] . ':' . $keys_value['08'] . "\n";

// Decoded arrays follow the runtime's regular nested value-copy behavior.
$copy = json_decode('{"nested":{"name":"source"}}', true);
$copy_alias = $copy;
$copy_nested = $copy['nested'];
$copy_nested['name'] = 'changed';
$copy['nested'] = $copy_nested;
echo 'copy:' . $copy['nested']['name'] . ':' . $copy_alias['nested']['name'] . "\n";

// Unicode escapes and an already-UTF-8 source string decode to byte strings.
// The NUL result is observed only by length so fixture output contains no NUL.
$unicode = json_decode('{"escaped":"\\u732b","pair":"\\uD83D\\uDE3A","raw":"猫","nul":"\\u0000"}', true);
echo 'unicode:' . $unicode['escaped'] . ':' . strlen($unicode['pair']) . ':' . strlen($unicode['raw']) . ':' . strlen($unicode['nul']) . "\n";

// Success -> malformed -> success proves error state publication and reset.
$first_valid = json_decode('{"ok":1}', true);
$broken = json_decode('{"broken":}', true);
echo 'reset-before:' . $first_valid['ok'] . ':' . (is_null($broken) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$after_valid = json_decode('{"ok":2}', true);
echo 'reset-after:' . $after_valid['ok'] . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";

// Distinct decoder errors stay observable without hardcoding their PHP values.
$mismatch = json_decode('{"object":1]', true);
echo 'mismatch:' . (is_null($mismatch) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$syntax = json_decode('{"broken":}', true);
echo 'syntax:' . (is_null($syntax) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$control_text = "{\"text\":\"a" . "\n" . "b\"}";
$control = json_decode($control_text, true);
echo 'control:' . (is_null($control) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$utf16 = json_decode('"\\uD800"', true);
echo 'utf16:' . (is_null($utf16) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$invalid_utf8_text = '"' . substr('猫', 0, 1) . '"';
$utf8 = json_decode($invalid_utf8_text, true);
echo 'utf8:' . (is_null($utf8) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$incomplete = json_decode('"open', true);
echo 'incomplete:' . (is_null($incomplete) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";
$bad_pair = json_decode('"\\uD800\\uZZZZ"', true);
echo 'bad-pair:' . (is_null($bad_pair) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";

// PHP's default depth is 512: 511 nested containers succeeds, while 512 is
// reported as a depth error. The loops avoid a gigantic source literal.
$depth_ok_text = '0';
$depth_index = 0;
while ($depth_index < 511) {
    $depth_ok_text = '[' . $depth_ok_text . ']';
    $depth_index = $depth_index + 1;
}
$depth_ok = json_decode($depth_ok_text, true);
$depth_bad_text = '0';
$depth_index = 0;
while ($depth_index < 512) {
    $depth_bad_text = '[' . $depth_bad_text . ']';
    $depth_index = $depth_index + 1;
}
$depth_bad = json_decode($depth_bad_text, true);
echo 'depth:' . count($depth_ok) . ':' . (is_null($depth_bad) ? 'null' : 'wrong') . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";

// The same Execution crosses user functions, methods, arrows, and array_map.
$function_error = decode_error_code('{"broken":}');
$probe = new JsonErrorProbe();
$method_error = $probe->decode_error_code('{"broken":}');
$arrow = fn($text) => json_decode($text, true);
$arrow_value = $arrow('{"arrow":true}');
$mapped = array_map('map_decode_value', ['{"id":1}', '{"id":2}']);
echo 'contexts:' . $function_error . ':' . $method_error . ':' . ($arrow_value['arrow'] ? 'true' : 'false') . ':' . $mapped[0]['id'] . ':' . $mapped[1]['id'] . ':' . json_last_error() . ':' . json_last_error_msg() . "\n";

// Invalid raw UTF-8 above is produced by a byte substring of valid UTF-8, using
// the existing admitted substr boundary; no new byte-construction builtin.

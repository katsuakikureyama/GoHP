<?php
// Trusted lexical sidecar. Input is data on stdin, never included or evaluated.
$source = stream_get_contents(STDIN, (16 << 20) + 1);
if ($source === false || strlen($source) > (16 << 20)) {
    exit(2);
}
$tokens = token_get_all($source, 0);
if (count($tokens) > 250000) {
    exit(2);
}
$observed = [];
foreach ($tokens as $token) {
    if (is_array($token)) {
        $kind = token_name($token[0]);
        $text = $token[1];
    } else {
        $kind = 'CHAR_' . strtoupper(bin2hex($token));
        $text = $token;
    }
    $observed[] = ['kind' => $kind, 'bytes' => strlen($text)];
}
$internal = get_defined_functions()['internal'];
$extensions = get_loaded_extensions();
$zendExtensions = get_loaded_extensions(true);
sort($internal, SORT_STRING);
sort($extensions, SORT_STRING);
sort($zendExtensions, SORT_STRING);
echo json_encode([
    'schema' => 'gohp.php.tokens/v0.2',
    'php_version' => PHP_VERSION,
    'bytes' => strlen($source),
    'sha256' => hash('sha256', $source),
    'tokens' => $observed,
    'internal_functions' => $internal,
    'extensions' => $extensions,
    'zend_extensions' => $zendExtensions,
], JSON_THROW_ON_ERROR);

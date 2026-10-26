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
echo json_encode([
    'schema' => 'gohp.php.tokens/v0.1',
    'php_version' => PHP_VERSION,
    'bytes' => strlen($source),
    'sha256' => hash('sha256', $source),
    'tokens' => $observed,
], JSON_THROW_ON_ERROR);

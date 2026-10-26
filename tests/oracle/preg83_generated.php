<?php
// Source-only B3 compiler fixture. The Go integration test appends its own
// observer so this file contains no PHP-side reporting or execution.

function preg83_basic() {
    $matches = null;
    $result = preg_match('/(a)(b)/s', 'zab', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_nomatch() {
    $matches = null;
    $result = preg_match('/z/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_named_unmatched() {
    $matches = null;
    $result = preg_match('/(?<letter>a)(b)?(c)?/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_negative_offset() {
    $matches = null;
    $result = preg_match('/a/s', 'xxa', $matches, PREG_OFFSET_CAPTURE, -1);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_middle_capture(string $subject, int $offset): array {
    $matches = null;
    $result = preg_match('/(a)(b)?(c)/s', $subject, $matches, PREG_OFFSET_CAPTURE, $offset);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_capture_copy() {
    $matches = null;
    $result = preg_match('/(?<letter>a)/s', 'a', $matches, PREG_OFFSET_CAPTURE, 0);
    $saved = $matches;
    $matches[0] = ['changed', 0];
    return ['saved' => $saved, 'changed' => $matches, 'result' => $result];
}

function preg83_binary_subject() {
    // The test turns the placeholder below into one literal NUL byte before both PHP and
    // GoHP parse the source; PHP single-quoted \x00 would not be a byte.
    $subject = 'left@NUL@right';
    $matches = null;
    $result = preg_match('/\\x00/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_utf8_byte_offset() {
    $subject = '雪a';
    $matches = null;
    $result = preg_match('/a/s', $subject, $matches, PREG_OFFSET_CAPTURE, 3);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_lookahead_possessive() {
    $matches = null;
    $result = preg_match('/a(?=b)b*+/s', 'abbb', $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_failure() {
    $subject = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!';
    $matches = null;
    $result = preg_match('/^(a+)+$/s', $subject, $matches, PREG_OFFSET_CAPTURE, 0);
    return ['result' => $result, 'matches' => $matches, 'error' => preg_last_error()];
}

function preg83_read_error() {
    return preg_last_error();
}

function preg83_make_reader() {
    return fn() => preg_last_error();
}

function preg83_invoke_reader(callable $reader) {
    return $reader();
}

class ErrorProbe {
    public function read_error() {
        return preg_last_error();
    }
}

function preg83_shared_execution() {
    $failed = preg83_failure();
    $reader = preg83_make_reader();
    $from_arrow = preg83_invoke_reader($reader);
    $probe = new ErrorProbe();
    $from_method = $probe->read_error();
    $reset = preg83_nomatch();
    return [
        'failure' => $failed,
        'arrow_error' => $from_arrow,
        'method_error' => $from_method,
        'reset' => $reset,
        'final_error' => preg_last_error(),
    ];
}

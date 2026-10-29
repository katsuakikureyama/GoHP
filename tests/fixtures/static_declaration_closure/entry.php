<?php
require_once __DIR__ . '/lib/first.php';
require_once __DIR__ . '/lib/second.php';
require_once __DIR__ . '/lib/first.php';

$shared = new SharedValue('seed', 0, '', false);
$alias = $shared;
$first = first_touch($alias);
$second = second_read($shared);
$other = new SharedValue('other', 0, '', false);
$other_read = second_read($other);
$falsy = (isset($shared->zero) ? '1' : '0') . (isset($shared->blank) ? '1' : '0') . (isset($shared->off) ? '1' : '0');
$source = ['tag' => 'source'];
$copy = $source;
$copy['tag'] = 'copy';

echo 'identity=' . $first . ':' . $second . ':' . $other_read . "\n";
echo 'falsy=' . $falsy . "\n";
echo 'copy=' . $source['tag'] . ':' . $copy['tag'] . "\n";

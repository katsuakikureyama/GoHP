<?php

/**
 * Native PHP reference/storage behavioral oracle.
 *
 * This file deliberately does not participate in GoHP differential testing.
 * It records the PHP 8.3.36 reference behavior that a later runtime/storage
 * implementation must reproduce before it can claim reference support.
 */

function reference_oracle_same(string $case, mixed $actual, mixed $expected): void
{
    if ($actual !== $expected) {
        throw new RuntimeException("case {$case} produced an unexpected value");
    }
}

/** @param list<array{id: string, status: string}> $cases */
function reference_oracle_case(string $id, callable $body, array &$cases): void
{
    try {
        $body();
    } catch (Throwable $error) {
        throw new RuntimeException("case {$id} failed: {$error->getMessage()}", 0, $error);
    }

    $cases[] = ['id' => $id, 'status' => 'pass'];
}

function reference_oracle_by_value_then_return(array $value): array
{
    $value['plain'] = 'returned';
    return $value;
}

function reference_oracle_rebind_argument(&$target): array
{
    $target = 'mutated';
    $replacement = 'replacement';
    $target =& $replacement;
    $target = 'rebound-local';

    return [$target, $replacement];
}

function &reference_oracle_return_reference(&$value)
{
    return $value;
}

$handlerInstalled = false;

try {
    if (PHP_MAJOR_VERSION !== 8 || PHP_MINOR_VERSION !== 3 || version_compare(PHP_VERSION, '8.3.36', '<')) {
        throw new RuntimeException("requires PHP 8.3 >= 8.3.36; got " . PHP_VERSION);
    }

    error_reporting(E_ALL);
    set_error_handler(static function (int $severity, string $message, string $file, int $line): bool {
        throw new ErrorException($message, 0, $severity, $file, $line);
    });
    $handlerInstalled = true;

    $cases = [];

    // A value assignment reads the current scalar; a reference assignment shares storage.
    reference_oracle_case('reference.scalar-alias-vs-value-copy', static function (): void {
        $value = 1;
        $alias =& $value;
        $copy = $value;
        $alias = 2;

        reference_oracle_same('reference.scalar-alias-vs-value-copy', [$value, $alias, $copy], [2, 2, 1]);
    }, $cases);

    // Rebinding the alias leaves the storage it used to name independent.
    reference_oracle_case('reference.scalar-rebinding', static function (): void {
        $left = 'left';
        $right = 'right';
        $alias =& $left;
        $alias =& $right;
        $alias = 'changed';

        reference_oracle_same('reference.scalar-rebinding', [$left, $right, $alias], ['left', 'changed', 'changed']);
    }, $cases);

    // Ordinary nested array assignment is a value copy at the observable boundary.
    reference_oracle_case('reference.nested-array-value-copy-isolation', static function (): void {
        $source = ['nested' => ['value' => 1]];
        $copy = $source;
        $copy['nested']['value'] = 2;

        reference_oracle_same('reference.nested-array-value-copy-isolation', [$source, $copy], [
            ['nested' => ['value' => 1]],
            ['nested' => ['value' => 2]],
        ]);
    }, $cases);

    // An explicit external reference remains live through an array value copy.
    reference_oracle_case('reference.copied-array-live-external-element', static function (): void {
        $external = 'seed';
        $source = ['ref' => &$external, 'plain' => 1];
        $copy = $source;
        $external = 'changed';

        reference_oracle_same('reference.copied-array-live-external-element', [$source['ref'], $copy['ref'], $external], [
            'changed',
            'changed',
            'changed',
        ]);
    }, $cases);

    // With no external alias left, the first element write after copying unshares that reference slot.
    reference_oracle_case('reference.array-copy-unsets-external-before-reference-write', static function (): void {
        $external = 1;
        $source = ['ref' => &$external, 'plain' => 0];
        $copy = $source;
        unset($external);
        $copy['ref'] = 2;

        reference_oracle_same('reference.array-copy-unsets-external-before-reference-write', [$source['ref'], $copy['ref']], [1, 2]);
    }, $cases);

    // Detaching a plain slot while an external alias is live preserves a shared reference slot.
    reference_oracle_case('reference.array-copy-detaches-plain-before-unset-external', static function (): void {
        $external = 1;
        $source = ['ref' => &$external, 'plain' => 0];
        $copy = $source;
        $copy['plain'] = 9;
        unset($external);
        $copy['ref'] = 2;

        reference_oracle_same('reference.array-copy-detaches-plain-before-unset-external', [$source['ref'], $copy['ref'], $source['plain'], $copy['plain']], [
            2,
            2,
            0,
            9,
        ]);
    }, $cases);

    // Passing and returning an array by value still preserves an element's explicit external reference.
    reference_oracle_case('reference.by-value-array-parameter-and-return-preserve-live-element', static function (): void {
        $external = 'before';
        $input = ['ref' => &$external, 'plain' => 'input'];
        $returned = reference_oracle_by_value_then_return($input);
        $external = 'after';

        reference_oracle_same('reference.by-value-array-parameter-and-return-preserve-live-element', [$input, $returned], [
            ['ref' => 'after', 'plain' => 'input'],
            ['ref' => 'after', 'plain' => 'returned'],
        ]);
    }, $cases);

    // Unsetting an alias removes that binding; recreating the name does not relink it.
    reference_oracle_case('reference.unset-variable-alias-detaches-name', static function (): void {
        $value = 1;
        $alias =& $value;
        unset($alias);
        $value = 2;
        $alias = 3;

        reference_oracle_same('reference.unset-variable-alias-detaches-name', [$value, $alias], [2, 3]);
    }, $cases);

    // Unsetting an element removes only that element binding, not the external storage it named.
    reference_oracle_case('reference.unset-element-alias-preserves-external-storage', static function (): void {
        $external = 1;
        $array = ['ref' => &$external];
        unset($array['ref']);
        $external = 2;
        $array['ref'] = 3;

        reference_oracle_same('reference.unset-element-alias-preserves-external-storage', [$external, $array], [2, ['ref' => 3]]);
    }, $cases);

    // A by-reference parameter can mutate the caller, while rebinding its local name is not a caller mutation.
    reference_oracle_case('reference.by-reference-argument-mutation-and-local-rebinding', static function (): void {
        $caller = 'before';
        $local = reference_oracle_rebind_argument($caller);

        reference_oracle_same('reference.by-reference-argument-mutation-and-local-rebinding', [$caller, $local], [
            'mutated',
            ['rebound-local', 'rebound-local'],
        ]);
    }, $cases);

    // A by-reference return exposes the original caller storage to a newly bound variable.
    reference_oracle_case('reference.by-reference-return-binding', static function (): void {
        $source = 'before';
        $returned =& reference_oracle_return_reference($source);
        $returned = 'after';

        reference_oracle_same('reference.by-reference-return-binding', [$source, $returned], ['after', 'after']);
    }, $cases);

    // foreach by reference keeps the final element bound until the loop variable is explicitly unset.
    reference_oracle_case('reference.foreach-final-binding-and-unset-detach', static function (): void {
        $items = [10, 20, 30];
        foreach ($items as &$item) {
        }
        $item = 99;
        unset($item);
        $item = 7;

        reference_oracle_same('reference.foreach-final-binding-and-unset-detach', [$items, $item], [[10, 20, 99], 7]);
    }, $cases);

    // Object assignment preserves object identity, whereas a variable reference can rebind the variable itself.
    reference_oracle_case('reference.object-identity-versus-variable-rebinding', static function (): void {
        $first = (object) ['name' => 'first'];
        $second = (object) ['name' => 'second'];
        $objectCopy = $first;
        $sameObjectBeforeRebind = $objectCopy === $first;
        $objectCopy->name = 'shared-mutation';
        $nameBeforeRebind = $first->name;
        $variableAlias =& $first;
        $variableAlias = $second;
        $objectCopy->name = 'copy-mutated';

        reference_oracle_same('reference.object-identity-versus-variable-rebinding', [
            $sameObjectBeforeRebind,
            $nameBeforeRebind,
            $first === $second,
            $objectCopy === $first,
            $first->name,
            $second->name,
            $objectCopy->name,
        ], [true, 'shared-mutation', true, false, 'second', 'second', 'copy-mutated']);
    }, $cases);

    // Taking a reference to a missing offset creates a present slot containing null before it is written.
    reference_oracle_case('reference.undefined-element-reference-creates-null-slot', static function (): void {
        $array = [];
        $created =& $array['missing'];
        $beforeWrite = [array_key_exists('missing', $array), $created];
        $created = 'created';

        reference_oracle_same('reference.undefined-element-reference-creates-null-slot', [$beforeWrite, $array], [
            [true, null],
            ['missing' => 'created'],
        ]);
    }, $cases);

    // Inspect a recursive array only through bounded scalar paths; never serialize or compare the full cycle.
    reference_oracle_case('reference.self-referential-array-bounded-observation', static function (): void {
        $node = ['value' => 1];
        $node['self'] =& $node;
        $node['self']['value'] = 2;

        reference_oracle_same('reference.self-referential-array-bounded-observation', [
            array_key_exists('self', $node),
            $node['value'],
            $node['self']['value'],
        ], [true, 2, 2]);
    }, $cases);

    $report = [
        'schema' => 'gohp.reference-oracle.v1',
        'scope' => 'native-oracle-only',
        'php_version' => PHP_VERSION,
        'case_count' => count($cases),
        'cases' => $cases,
    ];
    $encoded = json_encode($report, JSON_UNESCAPED_SLASHES);
    if ($encoded === false) {
        throw new RuntimeException('could not encode success report');
    }

    fwrite(STDOUT, $encoded . PHP_EOL);
    restore_error_handler();
    $handlerInstalled = false;
} catch (Throwable $error) {
    if ($handlerInstalled) {
        restore_error_handler();
    }
    fwrite(STDERR, 'reference oracle failure: ' . $error->getMessage() . PHP_EOL);
    exit(1);
}

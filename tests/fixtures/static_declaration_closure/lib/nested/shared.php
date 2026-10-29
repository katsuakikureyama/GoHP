<?php
class SharedValue {
    public $value;
    public $zero;
    public $blank;
    public $off;

    public function __construct($value, $zero, $blank, $off) {
        $this->value = $value;
        $this->zero = $zero;
        $this->blank = $blank;
        $this->off = $off;
    }
}

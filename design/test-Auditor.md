# Test Design: Auditor

**CRC Cards**: crc-Auditor.md
**Sequences**: seq-audit.md

### Test: ui-value on sl-badge (R34)
**Purpose**: Verify detection of ui-value bindings on sl-badge elements.

**Scenarios**:
1.  **Badge with ui-value**:
    - Create viewdef with `<sl-badge ui-value="count"></sl-badge>`.
    - Run audit.
    - Expect violation `ui_value_badge` with detail about using span inside badge.

2.  **Badge with span inside (Valid)**:
    - Create viewdef with `<sl-badge><span ui-value="count"></span></sl-badge>`.
    - Run audit.
    - Expect NO `ui_value_badge` violation.

3.  **Badge with other ui-* attributes (Valid)**:
    - Create viewdef with `<sl-badge ui-attr-variant="badgeVariant()">`.
    - Run audit.
    - Expect NO `ui_value_badge` violation.

### Test: Non-empty method args (R35)
**Purpose**: Verify detection of method calls with invalid arguments.

**Scenarios**:
1.  **Empty parens (Valid)**:
    - Create viewdef with `<span ui-value="getName()">`.
    - Run audit.
    - Expect NO `non_empty_method_args` violation.

2.  **Underscore placeholder (Valid)**:
    - Create viewdef with `<sl-input ui-value="setValue(_)">`.
    - Run audit.
    - Expect NO `non_empty_method_args` violation.

3.  **Invalid arg content**:
    - Create viewdef with `<span ui-value="getValue(x)">`.
    - Run audit.
    - Expect violation `non_empty_method_args` mentioning `getValue(x)`.

4.  **String literal in parens**:
    - Create viewdef with `<span ui-value="format('hello')">`.
    - Run audit.
    - Expect violation `non_empty_method_args`.

5.  **Multiple method calls, one invalid**:
    - Create viewdef with `<span ui-class-hidden="isValid()" ui-value="getData(arg)">`.
    - Run audit.
    - Expect one violation for `getData(arg)`.
    - Expect NO violation for `isValid()`.

### Test: Path syntax validation (R36)
**Purpose**: Verify full path syntax validation against grammar.

**Scenarios**:
1.  **Simple identifier (Valid)**:
    - Path: `name`
    - Expect valid.

2.  **Dotted path (Valid)**:
    - Path: `parent.child.value`
    - Expect valid.

3.  **Method call (Valid)**:
    - Path: `getName()`
    - Expect valid.

4.  **Method with underscore (Valid)**:
    - Path: `setValue(_)`
    - Expect valid.

5.  **Bracket accessor with number (Valid)**:
    - Path: `items[0].name`
    - Expect valid.

6.  **Bracket accessor with ident (Valid)**:
    - Path: `items[key].value`
    - Expect valid.

7.  **Path with query params (Valid)**:
    - Path: `items?wrapper=ViewList`
    - Expect valid.

8.  **Path with multiple query params (Valid)**:
    - Path: `search?keypress&wrapper=Input`
    - Expect valid.

9.  **Method in chain (Valid)**:
    - Path: `parent.getChild().name`
    - Expect valid.

10. **Double dot (Invalid)**:
    - Path: `foo..bar`
    - Expect violation `invalid_path_syntax`.

11. **Unclosed bracket (Invalid)**:
    - Path: `items[0.name`
    - Expect violation `invalid_path_syntax`.

12. **Empty path (Invalid)**:
    - Path: ``
    - Expect violation `invalid_path_syntax`.

13. **Leading dot (Invalid)**:
    - Path: `.name`
    - Expect violation `invalid_path_syntax`.

14. **Trailing dot (Invalid)**:
    - Path: `name.`
    - Expect violation `invalid_path_syntax`.

15. **ui-namespace excluded**:
    - Create viewdef with `<div ui-namespace="list-item">`.
    - Run audit.
    - Expect NO path syntax violation (ui-namespace is not a binding path).

### Test: Interaction with other checks
**Purpose**: Verify path checks work alongside existing checks.

**Scenarios**:
1.  **Operator check still works**:
    - Path: `!isHidden`
    - Expect `operator_in_path` violation (existing check).
    - May also get `invalid_path_syntax` (both checks apply).

2.  **item. prefix check still works**:
    - In list-item viewdef with `item.name`.
    - Expect `item_prefix` violation (existing check).
    - Path itself is syntactically valid.

3.  **Missing method check still works**:
    - Path: `unknownMethod()`
    - Expect `missing_method` violation if method not defined.
    - Path syntax is valid.

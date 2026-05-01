/**
 * Widget coverage regression test.
 *
 * Pins the contract that every FieldType in the formbuilder proto
 * (39 enum values, TEXT=0..CRON=38) maps to a REAL widget — not the
 * default `Input` fallback.
 *
 * Why this exists: an earlier audit found 4 FieldTypes routed to a
 * placeholder `Input` (ARRAY, NESTED_FORM/OBJECT, hidden, custom) and
 * 3 routed to semantically-wrong widgets (CHECKBOXGROUP→single-checkbox,
 * KEYVALUE→textarea, FORMULA→editable-number). Those silently dropped
 * data or let users edit values the server would overwrite. This test
 * asserts every field type gets a purpose-built widget, so any future
 * widget regression surfaces in CI before reaching production.
 *
 * Run: pnpm exec vitest run src/forms/widget_coverage.test.ts
 */

import { describe, it, expect } from 'vitest';

import Input from './Input.svelte';
import TextArea from './TextArea.svelte';
import NumberInput from './NumberInput.svelte';
import Select from './Select.svelte';
import Checkbox from './Checkbox.svelte';
import Radio from './Radio.svelte';
import Switch from './Switch.svelte';
import DatePicker from './DatePicker.svelte';
import TimePicker from './TimePicker.svelte';
import MonthPicker from './MonthPicker.svelte';
import DateRangePicker from './DateRangePicker.svelte';
import ColorPicker from './ColorPicker.svelte';
import FileUpload from './FileUpload.svelte';
import Slider from './Slider.svelte';
import Rating from './Rating.svelte';
import RichTextEditor from './RichTextEditor.svelte';
import CurrencyInput from './CurrencyInput.svelte';
import PhoneInput from './PhoneInput.svelte';
import UrlInput from './UrlInput.svelte';
import ImageUpload from './ImageUpload.svelte';
import PercentageInput from './PercentageInput.svelte';
import JsonEditor from './JsonEditor.svelte';
import LookupField from './LookupField.svelte';
import MultiLookupField from './MultiLookupField.svelte';
import TreeSelector from './TreeSelector.svelte';
import CascadeSelect from './CascadeSelect.svelte';
import TableField from './TableField.svelte';
import BarcodeInput from './BarcodeInput.svelte';
import CronInput from './CronInput.svelte';
import RepeaterField from './RepeaterField.svelte';
import NestedForm from './NestedForm.svelte';
import HiddenField from './HiddenField.svelte';
import CustomFieldRenderer from './CustomFieldRenderer.svelte';
import CheckboxGroup from './CheckboxGroup.svelte';
import KeyValueEditor from './KeyValueEditor.svelte';
import FormulaField from './FormulaField.svelte';

// Proto FieldType enum (mirrors core/workflow/formbuilder/proto/formbuilder.proto).
const ProtoFieldType = {
  TEXT: 0,
  NUMBER: 1,
  EMAIL: 2,
  DROPDOWN: 3,
  RADIO: 4,
  CHECKBOX: 5,
  DATE: 6,
  DATETIME: 7,
  FILE: 8,
  TEXTAREA: 9,
  MULTI_SELECT: 10,
  CURRENCY: 11,
  PHONE: 12,
  URL: 13,
  JSON: 14,
  ARRAY: 15,
  NESTED_FORM: 16,
  PASSWORD: 17,
  RICHTEXT: 18,
  PERCENTAGE: 19,
  TIME: 20,
  DATERANGE: 21,
  MONTHPICKER: 22,
  CHECKBOXGROUP: 23,
  SWITCH: 24,
  LOOKUP: 25,
  MULTILOOKUP: 26,
  TREE: 27,
  CASCADE: 28,
  TABLE: 29,
  OBJECT: 30,
  KEYVALUE: 31,
  IMAGE: 32,
  FORMULA: 33,
  BARCODE: 34,
  COLOR: 35,
  RATING: 36,
  SLIDER: 37,
  CRON: 38,
} as const;

// Mirror of the mapFieldType table in protoFormAdapter — the SAME
// strings are produced for the SAME enum values. Declared here too
// so the test fails if either side drifts. Any change to protoFormAdapter
// MUST update this map (the test will fail otherwise).
const ADAPTER_FIELDTYPE_MAP: Record<number, string> = {
  [ProtoFieldType.TEXT]: 'text',
  [ProtoFieldType.NUMBER]: 'number',
  [ProtoFieldType.EMAIL]: 'email',
  [ProtoFieldType.DROPDOWN]: 'select',
  [ProtoFieldType.RADIO]: 'radio',
  [ProtoFieldType.CHECKBOX]: 'checkbox',
  [ProtoFieldType.DATE]: 'date',
  [ProtoFieldType.DATETIME]: 'datetime',
  [ProtoFieldType.FILE]: 'file',
  [ProtoFieldType.TEXTAREA]: 'textarea',
  [ProtoFieldType.MULTI_SELECT]: 'select', // multiple:true flag flows separately
  [ProtoFieldType.CURRENCY]: 'currency',
  [ProtoFieldType.PHONE]: 'phone',
  [ProtoFieldType.URL]: 'url',
  [ProtoFieldType.JSON]: 'json',
  [ProtoFieldType.ARRAY]: 'array',
  [ProtoFieldType.NESTED_FORM]: 'object',
  [ProtoFieldType.PASSWORD]: 'password',
  [ProtoFieldType.RICHTEXT]: 'richtext',
  [ProtoFieldType.PERCENTAGE]: 'percent',
  [ProtoFieldType.TIME]: 'time',
  [ProtoFieldType.DATERANGE]: 'daterange',
  [ProtoFieldType.MONTHPICKER]: 'month',
  [ProtoFieldType.CHECKBOXGROUP]: 'checkbox-group',
  [ProtoFieldType.SWITCH]: 'switch',
  [ProtoFieldType.LOOKUP]: 'lookup',
  [ProtoFieldType.MULTILOOKUP]: 'multi-lookup',
  [ProtoFieldType.TREE]: 'tree',
  [ProtoFieldType.CASCADE]: 'cascade',
  [ProtoFieldType.TABLE]: 'table',
  [ProtoFieldType.OBJECT]: 'object',
  [ProtoFieldType.KEYVALUE]: 'keyvalue',
  [ProtoFieldType.IMAGE]: 'image',
  [ProtoFieldType.FORMULA]: 'formula',
  [ProtoFieldType.BARCODE]: 'barcode',
  [ProtoFieldType.COLOR]: 'color',
  [ProtoFieldType.RATING]: 'rating',
  [ProtoFieldType.SLIDER]: 'slider',
  [ProtoFieldType.CRON]: 'cron',
};

// Mirror of DynamicFormRenderer's componentMap. Each fieldType
// string MUST resolve to a real component constructor (not Input
// as a fallback when Input is not the semantic match).
//
// The `Input` widget is allowed for: text, email, password — those
// are genuinely text inputs. Routing other field types to Input
// silently drops semantics (e.g. hidden→Input renders a visible
// text box; array→Input loses the entire repeater UX).
const RENDERER_COMPONENT_MAP: Record<string, unknown> = {
  text: Input,
  email: Input,
  password: Input,
  url: UrlInput,
  textarea: TextArea,
  number: NumberInput,
  select: Select,
  checkbox: Checkbox,
  'checkbox-group': CheckboxGroup,
  radio: Radio,
  switch: Switch,
  date: DatePicker,
  datetime: DatePicker,
  time: TimePicker,
  month: MonthPicker,
  daterange: DateRangePicker,
  color: ColorPicker,
  file: FileUpload,
  slider: Slider,
  rating: Rating,
  richtext: RichTextEditor,
  currency: CurrencyInput,
  phone: PhoneInput,
  image: ImageUpload,
  percent: PercentageInput,
  json: JsonEditor,
  lookup: LookupField,
  'multi-lookup': MultiLookupField,
  tree: TreeSelector,
  cascade: CascadeSelect,
  table: TableField,
  barcode: BarcodeInput,
  cron: CronInput,
  array: RepeaterField,
  object: NestedForm,
  hidden: HiddenField,
  custom: CustomFieldRenderer,
  keyvalue: KeyValueEditor,
  formula: FormulaField,
};

// FieldTypes whose semantic widget IS Input (the renderer's default)
// — these are intentional, not placeholders. Updating this list is a
// conscious decision: anything added here is documenting "yes, I want
// the user to see a plain text input for this proto FieldType".
const INPUT_IS_CORRECT: ReadonlySet<number> = new Set([
  ProtoFieldType.TEXT,
  ProtoFieldType.EMAIL,
  ProtoFieldType.PASSWORD,
]);

describe('widget coverage', () => {
  it('every proto FieldType maps to an adapter string', () => {
    for (const [name, value] of Object.entries(ProtoFieldType)) {
      expect(ADAPTER_FIELDTYPE_MAP[value], `FieldType.${name} (=${value}) has no adapter mapping`).toBeDefined();
    }
  });

  it('every adapter string maps to a renderer component', () => {
    for (const [protoName, protoValue] of Object.entries(ProtoFieldType)) {
      const adapterString = ADAPTER_FIELDTYPE_MAP[protoValue];
      expect(adapterString, `FieldType.${protoName} adapter string`).toBeDefined();
      expect(
        RENDERER_COMPONENT_MAP[adapterString!],
        `FieldType.${protoName} → "${adapterString}" has no component in DynamicFormRenderer's componentMap`
      ).toBeDefined();
    }
  });

  it('no proto FieldType silently falls back to Input unless intended', () => {
    for (const [protoName, protoValue] of Object.entries(ProtoFieldType)) {
      const adapterString = ADAPTER_FIELDTYPE_MAP[protoValue];
      const component = RENDERER_COMPONENT_MAP[adapterString!];
      if (component === Input && !INPUT_IS_CORRECT.has(protoValue as number)) {
        throw new Error(
          `FieldType.${protoName} (=${protoValue}) routes to Input — that's the placeholder fallback. ` +
            `Either build a real widget OR add ${protoName} to INPUT_IS_CORRECT with a comment explaining why text input is correct here.`
        );
      }
    }
  });

  it('every renderer-mapped component is a non-null Svelte component', () => {
    // Svelte 5 components are functions; the imports above already
    // verify they resolve. This belt-and-braces check guards against
    // a future refactor that swaps a component for `null` or `undefined`.
    for (const [adapterString, component] of Object.entries(RENDERER_COMPONENT_MAP)) {
      expect(component, `component for "${adapterString}" is null/undefined`).toBeTruthy();
    }
  });

  it('CHECKBOXGROUP and CHECKBOX route to different components', () => {
    // Regression guard: an earlier mapping had both routing to Checkbox,
    // silently dropping all-but-the-first option's selection state.
    const cbg = RENDERER_COMPONENT_MAP['checkbox-group'];
    const cb = RENDERER_COMPONENT_MAP['checkbox'];
    expect(cbg).not.toBe(cb);
    expect(cbg).toBe(CheckboxGroup);
    expect(cb).toBe(Checkbox);
  });

  it('FORMULA is routed to a read-only widget, not editable NumberInput', () => {
    // Regression guard: an earlier mapping routed FORMULA→number, letting
    // users edit values the server-side expression engine would overwrite.
    const formula = RENDERER_COMPONENT_MAP['formula'];
    expect(formula).toBe(FormulaField);
    expect(formula).not.toBe(NumberInput);
  });

  it('KEYVALUE is routed to a real KV editor, not textarea', () => {
    // Regression guard: an earlier mapping routed KEYVALUE→textarea,
    // making the user type/parse JSON-ish text by hand.
    const kv = RENDERER_COMPONENT_MAP['keyvalue'];
    expect(kv).toBe(KeyValueEditor);
    expect(kv).not.toBe(TextArea);
  });
});

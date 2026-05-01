# Samavaya WASM Cheatsheet

> High-performance WebAssembly modules for Indian ERP operations

---

## Quick Usage Examples

### Import and Initialize

```typescript
// Import from the WASM package
import * as wasm from '@samavāya/wasm';

// Initialize specific modules (async, must await before using)
import initCore, * as core from '@samavāya/wasm/pkg/core/samavaya_core';
import initValidation, * as validation from '@samavāya/wasm/pkg/validation/samavaya_validation';
import initTaxEngine, * as taxEngine from '@samavāya/wasm/pkg/tax-engine/samavaya_tax_engine';
import initBarcode, * as barcode from '@samavāya/wasm/pkg/barcode/samavaya_barcode';
import initPayroll, * as payroll from '@samavāya/wasm/pkg/payroll/samavaya_payroll';

// Initialize before use
await initCore();
await initValidation();
await initTaxEngine();
await initBarcode();
await initPayroll();
```

### Using the Barrel Export (Recommended)

```typescript
// The index.ts provides namespaced exports
import { core, validation, taxEngine, barcode, payroll, ledger, pricing, bom, depreciation, compliance, crypto, i18n, offline } from '@samavāya/wasm';

// Each namespace requires initialization
await core.default();  // Initialize core module

// Then use functions directly
const formatted = core.format_indian_number('1234567');  // "12,34,567"
const words = core.amount_to_words('12345.50');          // "Twelve Thousand Three Hundred Forty Five and Fifty Paise"
```

---

## Real-World Usage Examples

### 1. Validate GSTIN with Details

```typescript
import { validation } from '@samavāya/wasm';

await validation.default();

const gstin = '27AAPFU0939F1ZV';

// Quick boolean check
const isValid = validation.validate_gstin(gstin);
console.log(isValid);  // true

// Parse GSTIN to get components
const parsed = validation.parse_gstin(gstin);
// {
//   gstin: "27AAPFU0939F1ZV",
//   valid: true,
//   state_code: "27",
//   state_name: "Maharashtra",
//   pan: "AAPFU0939F",
//   entity_code: "1",
//   entity_type: "Proprietorship",
//   checksum: "V"
// }

// Get specific parts
const pan = validation.get_pan_from_gstin(gstin);           // "AAPFU0939F"
const stateCode = validation.get_state_from_gstin(gstin);   // "27"

// Check if two GSTINs are from same state (for IGST determination)
const sameState = validation.is_same_state_gstin(
  '27AAPFU0939F1ZV',  // Maharashtra
  '27BBCDE1234F1Z5'   // Maharashtra
);  // true
```

### 2. Calculate GST with State Logic

```typescript
import { taxEngine } from '@samavāya/wasm';

await taxEngine.default();

// Inter-state sale (IGST)
const interStateGst = taxEngine.calculate_gst({
  amount: '10000',
  rate: '18',
  source_state: 'MH',    // Maharashtra (seller)
  dest_state: 'KA'       // Karnataka (buyer)
});
// {
//   taxable_amount: "10000",
//   cgst: "0",
//   sgst: "0",
//   igst: "1800",
//   total_tax: "1800",
//   total_amount: "11800"
// }

// Intra-state sale (CGST + SGST)
const intraStateGst = taxEngine.calculate_gst({
  amount: '10000',
  rate: '18',
  source_state: 'MH',
  dest_state: 'MH'
});
// {
//   taxable_amount: "10000",
//   cgst: "900",
//   sgst: "900",
//   igst: "0",
//   total_tax: "1800",
//   total_amount: "11800"
// }

// Simple convenience function
const totalTax = taxEngine.simple_gst('10000', '18', 'MH', 'KA');  // "1800"

// Reverse calculation (tax-inclusive price)
const fromInclusive = taxEngine.calculate_gst_inclusive(
  '11800',  // inclusive amount
  '18',     // GST rate
  'MH',     // source
  'KA'      // destination
);
// { taxable_amount: "10000", igst: "1800", ... }
```

### 3. TDS/TCS Calculation

```typescript
import { taxEngine } from '@samavāya/wasm';

await taxEngine.default();

// TDS calculation with section
const tds = taxEngine.calculate_tds({
  amount: '50000',
  section: '194A',      // Interest on deposits
  has_pan: true,
  is_resident: true
});
// {
//   gross_amount: "50000",
//   tds_rate: "10",
//   tds_amount: "5000",
//   net_amount: "45000",
//   section: "194A"
// }

// Without PAN (higher rate 20%)
const tdsNoPan = taxEngine.calculate_tds({
  amount: '50000',
  section: '194A',
  has_pan: false,
  is_resident: true
});
// { tds_rate: "20", tds_amount: "10000", ... }

// Simple TDS
const simpleTds = taxEngine.simple_tds('50000', '194A', true);  // "5000"

// TCS calculation
const tcs = taxEngine.calculate_tcs({
  amount: '1000000',
  section: '206C_1H',   // Sale of goods > 50L
  has_pan: true
});

// Get rate/threshold info
const tdsRate = taxEngine.get_tds_rate('194A');       // "10"
const threshold = taxEngine.get_tds_threshold('194A'); // "40000"
```

### 4. Validate Indian IDs

```typescript
import { validation } from '@samavāya/wasm';

await validation.default();

// PAN Validation
const isPanValid = validation.validate_pan('ABCDE1234F');  // true
const panType = validation.get_pan_holder_type('ABCDE1234F');  // "Individual"
const isCompany = validation.is_company_pan('ABCPC1234F');  // true (C = Company)

// Parse PAN
const panDetails = validation.parse_pan('ABCPC1234F');
// { pan: "ABCPC1234F", holder_type: "Company", valid: true }

// Aadhaar Validation (Verhoeff checksum)
const isAadhaarValid = validation.validate_aadhaar('123456789012');
const formattedAadhaar = validation.format_aadhaar('123456789012');  // "1234 5678 9012"
const maskedAadhaar = validation.mask_aadhaar('123456789012');       // "XXXX XXXX 9012"

// IFSC Validation
const isIfscValid = validation.validate_ifsc('SBIN0001234');
const bankDetails = validation.parse_ifsc('SBIN0001234');
// { ifsc: "SBIN0001234", bank_code: "SBIN", branch_code: "001234", bank_name: "State Bank of India" }

// CIN Validation
const isCinValid = validation.validate_cin('U74999MH2020PTC123456');
const cinDetails = validation.parse_cin('U74999MH2020PTC123456');
// { cin: "...", listing_status: "Unlisted", industry_code: "74999", state: "Maharashtra", incorporation_year: "2020", company_type: "Private Limited" }

// Mobile & Email
const isMobileValid = validation.validate_mobile('9876543210');  // true
const isEmailValid = validation.validate_email('test@example.com');  // true

// Password Strength
const passwordResult = validation.validate_password('MyP@ssw0rd!', 8);
// { valid: true, score: 4, feedback: [...] }
const strengthScore = validation.password_strength_score('MyP@ssw0rd!');  // 4 (out of 5)
```

### 5. Generate Barcodes & QR Codes

```typescript
import { barcode } from '@samavāya/wasm';

await barcode.default();

// Generate QR code (returns SVG string)
const qrSvg = barcode.generate_qr('https://example.com', 200);

// Generate QR as data URL (for img src)
const qrDataUrl = barcode.generate_qr_data_url('Hello World', 150);
// "data:image/svg+xml;base64,..."

// UPI Payment QR
const upiQr = barcode.generate_upi_qr(
  'merchant@upi',      // VPA
  'Merchant Name',     // Payee name
  1000,                // Amount (optional)
  'Order #123',        // Note (optional)
  250                  // Size (optional)
);

// GST Invoice QR (as per e-invoice spec)
const gstQr = barcode.generate_gst_invoice_qr(
  '27AAPFU0939F1ZV',   // Seller GSTIN
  '29AABCU9603R1ZM',   // Buyer GSTIN
  'INV-2024-001',      // Invoice number
  '2024-01-15',        // Invoice date
  '11800',             // Total value
  200                  // Size
);

// vCard QR for business cards
const vcardQr = barcode.generate_vcard_qr(
  'John Doe',
  '+919876543210',
  'john@example.com',
  'ABC Corp',
  200
);

// Code128 barcode (for SKUs, serial numbers)
const code128 = barcode.generate_code128('SKU-12345', 200, 50);

// EAN-13 barcode (for products)
const ean13 = barcode.generate_ean13('890123456789', 200, 80);

// EAN-8 barcode (for small products)
const ean8 = barcode.generate_ean8('1234567', 150, 60);

// Auto-calculate check digit
const withCheckDigit = barcode.ean13_with_check_digit('890123456789');  // "8901234567897"
```

### 6. Payroll Calculations

```typescript
import { payroll } from '@samavāya/wasm';

await payroll.default();

// Income Tax Calculation
const incomeTax = payroll.calculate_income_tax({
  gross_income: '1200000',    // 12 LPA
  regime: 'new',              // 'old' | 'new'
  age: 35,
  deductions: {
    section_80c: '150000',
    section_80d: '25000',
    hra_exemption: '0'        // Only for old regime
  }
});
// {
//   gross_income: "1200000",
//   taxable_income: "1050000",
//   tax_before_rebate: "...",
//   rebate_87a: "...",
//   tax_after_rebate: "...",
//   education_cess: "...",
//   total_tax: "...",
//   effective_rate: "...",
//   slab_breakdown: [...]
// }

// Compare Old vs New Regime
const comparison = payroll.compare_tax_regimes({
  gross_income: '1500000',
  age: 35,
  deductions: {
    section_80c: '150000',
    section_80d: '50000',
    hra_exemption: '300000'
  }
});
// {
//   old_regime: { total_tax: "...", effective_rate: "..." },
//   new_regime: { total_tax: "...", effective_rate: "..." },
//   recommendation: "new",
//   savings: "45000"
// }

// Get tax slabs
const newSlabs = payroll.get_tax_slabs('new', '2024-25');

// PF Calculation
const pf = payroll.calculate_pf({
  basic_salary: '35000',
  da: '0',
  pf_on_full_basic: false
});
// {
//   wage_for_pf: "15000",  // Ceiling
//   employee_contribution: "1800",
//   employer_pf_contribution: "550",    // 3.67%
//   employer_eps_contribution: "1250",  // 8.33%
//   employer_total: "1800"
// }

// ESI Calculation
const esi = payroll.calculate_esi({
  gross_salary: '18000',
  state: 'MH'
});
// {
//   is_applicable: true,   // Under 21000 ceiling
//   employee_contribution: "135",  // 0.75%
//   employer_contribution: "585",  // 3.25%
// }

// Professional Tax
const pt = payroll.calculate_professional_tax({
  gross_salary: '50000',
  state: 'MH'
});
// { monthly_tax: "200", annual_tax: "2400" }

// All Statutory Deductions at Once
const allStatutory = payroll.calculate_all_statutory(
  '40000',   // Basic salary
  '50000',   // Gross salary
  'MH',      // State
  '0'        // DA (optional)
);
// {
//   pf: {...},
//   esi: {...},
//   professional_tax: {...},
//   lwf: {...},
//   total_employee_deduction: "...",
//   total_employer_contribution: "..."
// }

// CTC Breakdown
const ctc = payroll.calculate_ctc_breakdown({
  annual_ctc: '1200000',
  include_gratuity: true,
  pf_on_full_basic: false,
  include_bonus: true,
  metro_city: true
});
// {
//   annual_ctc: "1200000",
//   basic: "480000",
//   hra: "192000",
//   special_allowance: "...",
//   employer_pf: "...",
//   gratuity: "...",
//   monthly_gross: "...",
//   annual_gross: "..."
// }

// Optimize Salary Structure
const optimized = payroll.optimize_salary_structure('1200000', 'MH', true);

// Reverse CTC (from in-hand to CTC)
const reverseCTC = payroll.reverse_ctc_calculation('80000', 'MH', false);
```

### 7. Indian Number Formatting

```typescript
import { core } from '@samavāya/wasm';

await core.default();

// Indian number system (lakhs, crores)
const formatted = core.format_indian_number('12345678');
// "1,23,45,678"

// Amount to words
const words = core.amount_to_words('123456.50');
// "One Lakh Twenty Three Thousand Four Hundred Fifty Six and Fifty Paise"

// Currency formatting
const money = core.format_money('123456.50', 'INR');
// "₹1,23,456.50"

// Decimal operations (for precise financial math)
const sum = core.add_decimals('10.50', '20.75');      // "31.25"
const diff = core.subtract_decimals('100', '30.50');  // "69.50"
const product = core.multiply_decimals('100', '0.18'); // "18.00"
const quotient = core.divide_decimals('100', '3');     // "33.33..."
const pct = core.percentage('1000', '18');             // "180"
const rounded = core.round_currency('123.456');        // "123.46"

// Financial year info
const fy = core.get_financial_year('2024-01-15');
// { year: "2023-24", start_date: "2023-04-01", end_date: "2024-03-31" }

const currentFY = core.get_current_financial_year();
const quarter = core.get_quarter_from_date('2024-01-15');  // "Q4"

// State utilities
const allStates = core.get_all_states();
const state = core.get_state_by_code('MH');
// { code: "MH", name: "Maharashtra", gst_code: "27", is_ut: false }

const stateFromGst = core.get_state_by_gst_code('27');  // Maharashtra
const isUT = core.is_union_territory('DD');  // true (Daman & Diu)
const isIntraState = core.is_intra_state('MH', 'MH');  // true
```

### 8. HSN/SAC Lookup

```typescript
import { taxEngine } from '@samavāya/wasm';

await taxEngine.default();

// Validate HSN code
const isValidHsn = taxEngine.validate_hsn('84713010');  // true

// Validate SAC code
const isValidSac = taxEngine.validate_sac('998311');  // true

// Get GST rate for HSN
const hsnRate = taxEngine.get_hsn_gst_rate('84713010');  // "18"

// Get GST rate for SAC
const sacRate = taxEngine.get_sac_gst_rate('998311');  // "18"

// Full lookup
const hsnInfo = taxEngine.lookup_hsn_sac('84713010');
// { code: "84713010", description: "Portable computers", gst_rate: "18", type: "goods" }

// Check if service code
const isService = taxEngine.is_service_code('998311');  // true

// Get chapter from HSN
const chapter = taxEngine.get_hsn_chapter('84713010');  // "84"

// Cess info
const hasCess = taxEngine.has_cess('27101990');  // true (petrol)
const cessRate = taxEngine.get_cess_rate('27101990');
```

### 9. Using in Svelte Components

```svelte
<script lang="ts">
  import { validation, taxEngine, core, barcode } from '@samavāya/wasm';
  import { onMount } from 'svelte';

  let gstin = $state('');
  let gstinResult = $state<any>(null);
  let isLoading = $state(true);

  onMount(async () => {
    // Initialize WASM modules
    await Promise.all([
      validation.default(),
      taxEngine.default(),
      core.default(),
      barcode.default()
    ]);
    isLoading = false;
  });

  function validateGstin() {
    if (gstin.length === 15) {
      gstinResult = validation.parse_gstin(gstin);
    }
  }
</script>

{#if isLoading}
  <p>Loading WASM modules...</p>
{:else}
  <div class="gstin-validator">
    <input
      bind:value={gstin}
      oninput={validateGstin}
      placeholder="Enter GSTIN"
      maxlength="15"
    />

    {#if gstinResult}
      {#if gstinResult.valid}
        <div class="valid">
          ✓ Valid GSTIN
          <p>State: {gstinResult.state_name}</p>
          <p>PAN: {gstinResult.pan}</p>
          <p>Entity Type: {gstinResult.entity_type}</p>
        </div>
      {:else}
        <div class="invalid">✗ Invalid GSTIN</div>
      {/if}
    {/if}
  </div>
{/if}
```

### 10. Invoice Calculator Component

```svelte
<script lang="ts">
  import { taxEngine, core } from '@samavāya/wasm';
  import { onMount } from 'svelte';

  let ready = $state(false);
  let amount = $state('10000');
  let gstRate = $state('18');
  let sourceState = $state('MH');
  let destState = $state('KA');
  let result = $state<any>(null);

  onMount(async () => {
    await Promise.all([taxEngine.default(), core.default()]);
    ready = true;
    calculate();
  });

  function calculate() {
    if (!ready) return;

    const gst = taxEngine.calculate_gst({
      amount,
      rate: gstRate,
      source_state: sourceState,
      dest_state: destState
    });

    result = {
      ...gst,
      formatted_total: core.format_money(gst.total_amount, 'INR'),
      amount_in_words: core.amount_to_words(gst.total_amount)
    };
  }
</script>

{#if ready}
  <div class="invoice-calc">
    <input type="text" bind:value={amount} oninput={calculate} />
    <select bind:value={gstRate} onchange={calculate}>
      <option value="5">5%</option>
      <option value="12">12%</option>
      <option value="18">18%</option>
      <option value="28">28%</option>
    </select>

    {#if result}
      <table>
        <tr><td>Taxable Amount</td><td>₹{result.taxable_amount}</td></tr>
        {#if result.igst !== '0'}
          <tr><td>IGST ({gstRate}%)</td><td>₹{result.igst}</td></tr>
        {:else}
          <tr><td>CGST ({Number(gstRate)/2}%)</td><td>₹{result.cgst}</td></tr>
          <tr><td>SGST ({Number(gstRate)/2}%)</td><td>₹{result.sgst}</td></tr>
        {/if}
        <tr class="total"><td>Total</td><td>{result.formatted_total}</td></tr>
      </table>
      <p class="words">{result.amount_in_words}</p>
    {/if}
  </div>
{/if}
```

---

## Implementation Status

### ✅ All Modules Implemented

| Crate | Module | Status | Functions |
|-------|--------|--------|-----------|
| `core` | Shared utilities | ✅ Complete | States, number formatting, decimals |
| `tax-engine` | Tax calculations | ✅ Complete | GST, TDS, TCS, HSN, Cess |
| `validation` | ID validation | ✅ Complete | GSTIN, PAN, TAN, CIN, IFSC, Aadhaar |
| `barcode` | Barcode generation | ✅ Complete | QR, UPI, Code128, EAN |
| `ledger` | Accounting logic | ✅ Complete | Journal, Trial Balance, Aging, Reconciliation |
| `pricing` | Price calculations | ✅ Complete | Discounts, Margins, Tiered pricing |
| `payroll` | Salary processing | ✅ Complete | Tax slabs, PF/ESI, CTC breakdown |
| `bom` | Bill of Materials | ✅ Complete | Cost rollup, explosion, where-used, MRP |
| `depreciation` | Asset depreciation | ✅ Complete | SLM, WDV, DDB, SYD, UOP, IT Act rates |
| `compliance` | Regulatory | ✅ Complete | e-Invoice, GSTR-1, GSTR-3B, E-Way Bill |
| `crypto` | Encryption | ✅ Complete | SHA-256/512, HMAC, PBKDF2, AES-GCM |
| `i18n` | Localization | ✅ Complete | Indian number formats, currency, dates, words |
| `offline` | Offline support | ✅ Complete | Sync, conflict resolution, version vectors |

---

## Full WASM Inventory by Domain

### Identity & Auth
| Component | Status | Description |
|-----------|--------|-------------|
| `validation/password` | ✅ | Password strength validation |
| `validation/aadhaar` | ✅ | Aadhaar Verhoeff validation |
| `crypto/hash` | ✅ | SHA-256, SHA-512, HMAC |
| `crypto/encrypt` | ✅ | AES-GCM encryption |
| `crypto/password` | ✅ | PBKDF2 password hashing |

### Masters
| Component | Status | Description |
|-----------|--------|-------------|
| `validation/gstin` | ✅ | GSTIN validation with checksum |
| `validation/pan` | ✅ | PAN validation & holder type |
| `validation/ifsc` | ✅ | IFSC with bank lookup |
| `validation/hsn` | ✅ | HSN/SAC validation |
| `barcode/*` | ✅ | QR, Code128, EAN barcodes |

### Finance
| Component | Status | Description |
|-----------|--------|-------------|
| `tax-engine/gst` | ✅ | GST calculation (CGST/SGST/IGST) |
| `tax-engine/tds` | ✅ | TDS deduction |
| `tax-engine/tcs` | ✅ | TCS collection |
| `ledger/journal` | ✅ | Journal validation |
| `ledger/trial-balance` | ✅ | Trial balance computation |
| `ledger/aging` | ✅ | AR/AP aging analysis |
| `ledger/reconciliation` | ✅ | Bank reconciliation |
| `depreciation/slm` | ✅ | Straight-line method |
| `depreciation/wdv` | ✅ | Written-down value |
| `depreciation/ddb` | ✅ | Double declining balance |
| `depreciation/syd` | ✅ | Sum of years digits |
| `depreciation/uop` | ✅ | Units of production |

### Sales & Purchase
| Component | Status | Description |
|-----------|--------|-------------|
| `pricing/calculate` | ✅ | Price with discounts |
| `pricing/margin` | ✅ | Margin/markup calculations |
| `pricing/tiered` | ✅ | Quantity-based pricing |
| `compliance/einvoice` | ✅ | E-invoice JSON generation |
| `compliance/gstr1` | ✅ | GSTR-1 generation & validation |
| `compliance/gstr3b` | ✅ | GSTR-3B computation |
| `compliance/eway` | ✅ | E-Way Bill generation |

### Inventory & Manufacturing
| Component | Status | Description |
|-----------|--------|-------------|
| `bom/explosion` | ✅ | Multi-level BOM explosion |
| `bom/costing` | ✅ | BOM cost rollup |
| `bom/where-used` | ✅ | Where-used analysis |
| `bom/mrp` | ✅ | Material requirements planning |
| `bom/eoq` | ✅ | Economic order quantity |

### HR & Payroll
| Component | Status | Description |
|-----------|--------|-------------|
| `payroll/income-tax` | ✅ | Income tax calculation (Old/New regime) |
| `payroll/pf` | ✅ | Provident Fund computation |
| `payroll/esi` | ✅ | ESI computation |
| `payroll/pt` | ✅ | Professional Tax (all states) |
| `payroll/lwf` | ✅ | Labour Welfare Fund |
| `payroll/ctc` | ✅ | CTC breakdown & optimization |

### Documents
| Component | Status | Description |
|-----------|--------|-------------|
| `barcode/qr` | ✅ | QR code generation |
| `barcode/upi` | ✅ | UPI payment QR |
| `barcode/gst-invoice` | ✅ | GST e-invoice QR |

### Platform
| Component | Status | Description |
|-----------|--------|-------------|
| `core/indian` | ✅ | Indian states, number format |
| `core/decimal` | ✅ | Precise decimal arithmetic |
| `i18n/numbers` | ✅ | Indian number formatting (lakhs/crores) |
| `i18n/currency` | ✅ | Currency formatting with symbols |
| `i18n/dates` | ✅ | Date formatting & financial year |
| `i18n/words` | ✅ | Amount to words (English/Hindi) |
| `offline/sync` | ✅ | Offline data sync |
| `offline/conflict` | ✅ | Conflict resolution |
| `offline/version-vector` | ✅ | Distributed versioning |

---

## Quick Start

```bash
# Install Rust and wasm-pack (one-time setup)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
cargo install wasm-pack

# Build all WASM modules
pnpm wasm:build

# Build in development mode (faster, larger output)
pnpm wasm:build:dev

# Clean build artifacts
pnpm wasm:clean
```

---

## Module Overview

| Module | Import | Description |
|--------|--------|-------------|
| `tax` | `@samavāya/wasm/tax` | GST, TDS, TCS, HSN |
| `validation` | `@samavāya/wasm/validation` | GSTIN, PAN, Aadhaar, IFSC |
| `barcode` | `@samavāya/wasm/barcode` | QR, Code128, EAN |
| `ledger` | `@samavāya/wasm/ledger` | Journal, Trial Balance, Aging |
| `pricing` | `@samavāya/wasm/pricing` | Discounts, Margins |
| `core` | `@samavāya/wasm/core` | Indian utilities |
| `payroll` | `@samavāya/wasm/payroll` | Income tax, PF, ESI, CTC |
| `bom` | `@samavāya/wasm/bom` | BOM explosion, MRP, costing |
| `depreciation` | `@samavāya/wasm/depreciation` | Asset depreciation |
| `compliance` | `@samavāya/wasm/compliance` | e-Invoice, GSTR |
| `crypto` | `@samavāya/wasm/crypto` | Hashing, encryption |
| `i18n` | `@samavāya/wasm/i18n` | Localization |
| `offline` | `@samavāya/wasm/offline` | Sync, conflicts |

---

## Tax Engine

### GST Calculation

```typescript
import { calculateGst } from '@samavāya/wasm';

// Inter-state (IGST)
const result = await calculateGst({
  amount: '10000',
  rate: '18',
  isInclusive: false,
  sourceState: 'MH',    // Maharashtra
  destState: 'KA'       // Karnataka
});
// → { igst: '1800', cgst: '0', sgst: '0', totalAmount: '11800' }

// Intra-state (CGST + SGST)
const result2 = await calculateGst({
  amount: '10000',
  rate: '18',
  isInclusive: false,
  sourceState: 'MH',
  destState: 'MH'
});
// → { cgst: '900', sgst: '900', igst: '0', totalAmount: '11800' }

// Tax-inclusive price
const result3 = await calculateGst({
  amount: '11800',
  rate: '18',
  isInclusive: true,
  sourceState: 'MH',
  destState: 'MH'
});
// → { taxableAmount: '10000', cgst: '900', sgst: '900' }
```

### TDS/TCS Calculation

```typescript
import { calculateTds, calculateTcs } from '@samavāya/wasm';

// Section 194A - Interest (Bank)
const tds = await calculateTds('194A', '50000', true);
// → { tdsAmount: '5000', tdsRate: '10', netAmount: '45000' }

// Without PAN (higher rate)
const tdsNoPan = await calculateTds('194A', '50000', false);
// → { tdsAmount: '10000', tdsRate: '20', netAmount: '40000' }

// TCS
const tcs = await calculateTcs('206C_1H', '1000000');
// → { tcsAmount: '1000', tcsRate: '0.1', totalAmount: '1001000' }
```

---

## Payroll

### Income Tax Calculation

```typescript
import { calculateIncomeTax, compareTaxRegimes } from '@samavāya/wasm';

// Calculate income tax
const tax = await calculateIncomeTax({
  grossIncome: '1200000',      // 12 LPA
  regime: 'new',               // 'old' | 'new'
  age: 35,
  deductions: {
    section80C: '150000',
    section80D: '25000',
    hra: '240000',
    homeLoanInterest: '200000'
  }
});
// → {
//     taxableIncome: '1050000',
//     totalTax: '54600',
//     effectiveRate: '4.55',
//     slabWiseBreakdown: [...]
//   }

// Compare regimes
const comparison = await compareTaxRegimes('1500000', {
  section80C: '150000',
  section80D: '50000',
  hra: '300000'
});
// → {
//     oldRegime: { totalTax: '195000' },
//     newRegime: { totalTax: '150000' },
//     recommendation: 'new',
//     savings: '45000'
//   }
```

### Statutory Deductions (PF, ESI, PT)

```typescript
import {
  calculatePf,
  calculateEsi,
  calculateProfessionalTax,
  calculateAllStatutory
} from '@samavāya/wasm';

// PF Calculation
const pf = await calculatePf('35000', false);
// → {
//     employeeContribution: '1800',  // 12% of 15000 (ceiling)
//     employerContribution: '1800',
//     employerPensionContribution: '1250',  // 8.33% to EPS
//     employerPfContribution: '550',        // 3.67% to EPF
//     wageForPf: '15000',
//     isAboveCeiling: true
//   }

// ESI Calculation
const esi = await calculateEsi('18000', 'MH');
// → {
//     employeeContribution: '135',   // 0.75%
//     employerContribution: '585',   // 3.25%
//     isApplicable: true             // Under 21000 ceiling
//   }

// Professional Tax (state-wise)
const pt = await calculateProfessionalTax('50000', 'MH');
// → { monthlyTax: '200', annualTax: '2400', state: 'MH' }

const ptKA = await calculateProfessionalTax('50000', 'KA');
// → { monthlyTax: '200', annualTax: '2400', state: 'KA' }

// All statutory at once
const statutory = await calculateAllStatutory('50000', '25000', 'MH', false);
// → {
//     pf: {...},
//     esi: {...},
//     professionalTax: {...},
//     lwf: {...},
//     totalEmployeeDeduction: '2135',
//     totalEmployerContribution: '2385'
//   }
```

### CTC Breakdown

```typescript
import {
  calculateCtcBreakdown,
  optimizeSalaryStructure,
  calculateSalaryStructure
} from '@samavāya/wasm';

// Break down CTC
const ctc = await calculateCtcBreakdown('1200000', {
  includeGratuity: true,
  pfOnFullBasic: false,
  includeBonus: true
});
// → {
//     annualCtc: '1200000',
//     basic: '480000',      // 40%
//     hra: '192000',        // 40% of basic
//     specialAllowance: '328000',
//     employerPf: '57600',
//     gratuity: '23076',
//     monthlyGross: '93333',
//     annualGross: '1120000'
//   }

// Optimize for tax savings
const optimized = await optimizeSalaryStructure('1200000');
// → {
//     originalStructure: {...},
//     optimizedStructure: {...},
//     taxSavings: '45000',
//     recommendations: [
//       'Increase HRA component',
//       'Add LTA allowance',
//       'Consider NPS contribution'
//     ]
//   }

// Monthly salary with proration
const monthly = await calculateSalaryStructure(ctc, 26, 22);
// → {
//     grossSalary: '93333',
//     proratedAmount: '78974',  // 22/26 days
//     deductions: [...],
//     netSalary: '76000'
//   }
```

---

## BOM (Bill of Materials)

### BOM Explosion

```typescript
import { explodeBom, explodeBomSingleLevel, validateBomCircular } from '@samavāya/wasm';

const bomData = {
  'FG-001': [
    { componentId: 'SA-001', componentName: 'Sub-Assembly', quantity: '2', uom: 'NOS' },
    { componentId: 'RM-001', componentName: 'Raw Material 1', quantity: '0.5', uom: 'KG' }
  ],
  'SA-001': [
    { componentId: 'RM-002', componentName: 'Raw Material 2', quantity: '0.25', uom: 'KG' },
    { componentId: 'RM-003', componentName: 'Fastener', quantity: '3', uom: 'NOS' }
  ]
};

// Multi-level explosion
const explosion = await explodeBom('FG-001', bomData, '100');
// → {
//     productId: 'FG-001',
//     totalLevels: 2,
//     items: [
//       { componentId: 'SA-001', level: 1, requiredQuantity: '200' },
//       { componentId: 'RM-001', level: 1, requiredQuantity: '50' },
//       { componentId: 'RM-002', level: 2, requiredQuantity: '50' },
//       { componentId: 'RM-003', level: 2, requiredQuantity: '600' }
//     ],
//     summary: {
//       'RM-001': { totalQuantity: '50', uom: 'KG' },
//       'RM-002': { totalQuantity: '50', uom: 'KG' },
//       'RM-003': { totalQuantity: '600', uom: 'NOS' }
//     }
//   }

// Validate for circular references
const validation = await validateBomCircular('FG-001', bomData);
// → { isValid: true }
```

### BOM Costing

```typescript
import { calculateCostRollup, calculateBomCost, makeVsBuyAnalysis } from '@samavāya/wasm';

const itemCosts = {
  'RM-001': '500',   // per KG
  'RM-002': '800',   // per KG
  'RM-003': '10'     // per NOS
};

// Cost rollup
const costs = await calculateCostRollup('FG-001', bomData, itemCosts);
// → {
//     productId: 'FG-001',
//     materialCost: '420',
//     totalCost: '420',
//     componentCosts: [...]
//   }

// Total cost for quantity
const batchCost = await calculateBomCost('FG-001', bomData, '100', itemCosts);
// → { unitCost: '420', totalCost: '42000' }

// Make vs Buy analysis
const analysis = await makeVsBuyAnalysis('FG-001', bomData, itemCosts, {
  'FG-001': '400'  // Buy price
});
// → { makeCost: '420', buyCost: '400', recommendation: 'buy', savings: '20' }
```

### MRP (Material Requirements Planning)

```typescript
import { runMrp, calculateEoq, calculateSafetyStock } from '@samavāya/wasm';

// Run MRP
const requirements = [
  { productId: 'FG-001', quantity: '100', dueDate: '2024-02-15' },
  { productId: 'FG-001', quantity: '50', dueDate: '2024-02-28' }
];

const inventory = { 'RM-001': '20', 'RM-002': '10', 'RM-003': '100' };
const inTransit = { 'RM-001': '10' };

const mrp = await runMrp(requirements, bomData, inventory, inTransit);
// → {
//     plannedOrders: [
//       { itemId: 'RM-001', quantity: '45', dueDate: '2024-02-10', orderType: 'purchase' },
//       { itemId: 'RM-002', quantity: '65', dueDate: '2024-02-10', orderType: 'purchase' },
//       ...
//     ],
//     shortages: [],
//     projectedInventory: {...}
//   }

// Calculate EOQ
const eoq = await calculateEoq('12000', '50', '2');
// → { eoq: '548', annualOrderingCost: '1095', annualHoldingCost: '548' }

// Calculate safety stock
const safetyStock = await calculateSafetyStock('100', '20', 7, '0.95');
// → '87'
```

---

## Depreciation

### Depreciation Methods

```typescript
import {
  calculateDepreciationSchedule,
  calculateSlm,
  calculateWdv,
  calculateDdb,
  getItDepreciationRates,
  compareDepreciationMethods
} from '@samavāya/wasm';

// Straight Line Method
const slm = await calculateSlm('1000000', '100000', 10);
// → {
//     method: 'SLM',
//     entries: [
//       { year: 1, openingValue: '1000000', depreciationAmount: '90000', closingValue: '910000' },
//       { year: 2, openingValue: '910000', depreciationAmount: '90000', closingValue: '820000' },
//       ...
//     ]
//   }

// Written Down Value (IT Act)
const wdv = await calculateWdv('1000000', '15', 10);
// → {
//     entries: [
//       { year: 1, openingValue: '1000000', depreciationAmount: '150000', closingValue: '850000' },
//       { year: 2, openingValue: '850000', depreciationAmount: '127500', closingValue: '722500' },
//       ...
//     ]
//   }

// Get IT Act rates
const rates = await getItDepreciationRates();
// → {
//     'buildings': { rate: '10', description: 'Buildings', examples: ['Factory', 'Office'] },
//     'furniture': { rate: '10', description: 'Furniture & Fittings' },
//     'plant_machinery': { rate: '15', description: 'Plant & Machinery' },
//     'computers': { rate: '40', description: 'Computers & Software' },
//     'vehicles': { rate: '15', description: 'Motor Vehicles' },
//     'intangibles': { rate: '25', description: 'Intangible Assets' }
//   }

// Compare all methods
const comparison = await compareDepreciationMethods('1000000', '100000', 10);
// → {
//     SLM: { totalDepreciation: '900000', entries: [...] },
//     WDV: { totalDepreciation: '900000', entries: [...] },
//     DDB: { totalDepreciation: '900000', entries: [...] },
//     SYD: { totalDepreciation: '900000', entries: [...] }
//   }
```

### Block Depreciation (IT Act)

```typescript
import { calculateBlockDepreciation } from '@samavāya/wasm';

const block = {
  blockId: 'PLANT-15',
  blockName: 'Plant & Machinery',
  rate: '15',
  openingWdv: '1000000',
  assets: [...]
};

const result = await calculateBlockDepreciation(
  block,
  [{ cost: '200000', date: '2024-06-15' }],   // additions
  [{ saleValue: '50000', date: '2024-10-01', originalCost: '100000' }],  // disposals
  '2024-25'
);
// → {
//     openingWdv: '1000000',
//     additions: '200000',
//     disposals: '50000',
//     baseForDepreciation: '1150000',
//     depreciationAmount: '172500',
//     closingWdv: '977500',
//     shortTermGain: '0',
//     longTermGain: '0'
//   }
```

---

## Compliance

### E-Invoice Generation

```typescript
import { generateEinvoiceJson, validateEinvoice, generateIrnHash } from '@samavāya/wasm';

const einvoice = await generateEinvoiceJson({
  version: '1.1',
  tranDtls: { taxSch: 'GST', supTyp: 'B2B', regRev: 'N', igstOnIntra: 'N' },
  docDtls: { typ: 'INV', no: 'INV-2024-001', dt: '15/01/2024' },
  sellerDtls: {
    gstin: '27AAPFU0939F1ZV',
    lglNm: 'ABC Traders Pvt Ltd',
    addr1: '123 Main Street',
    loc: 'Mumbai',
    pin: '400001',
    stcd: '27'
  },
  buyerDtls: {
    gstin: '29AABCU9603R1ZM',
    lglNm: 'XYZ Enterprises',
    pos: '29',
    addr1: '456 MG Road',
    loc: 'Bangalore',
    pin: '560001',
    stcd: '29'
  },
  itemList: [
    {
      slNo: '1',
      prdDesc: 'Laptop Computer',
      isServc: 'N',
      hsnCd: '84713010',
      qty: '10',
      unit: 'NOS',
      unitPrice: '50000',
      totAmt: '500000',
      discount: '0',
      preTaxVal: '500000',
      assAmt: '500000',
      gstRt: '18',
      igstAmt: '90000',
      cgstAmt: '0',
      sgstAmt: '0',
      cesRt: '0',
      cesAmt: '0',
      totItemVal: '590000'
    }
  ],
  valDtls: {
    assVal: '500000',
    cgstVal: '0',
    sgstVal: '0',
    igstVal: '90000',
    cesVal: '0',
    totInvVal: '590000'
  }
});
// → { json: '...', valid: true, errors: [], warnings: [] }

// Validate e-invoice JSON
const validation = await validateEinvoice(einvoice.json);
// → { valid: true, errors: [], warnings: [] }

// Generate IRN hash
const irn = await generateIrnHash('INV-2024-001', '27AAPFU0939F1ZV', '2024-25');
// → 'a1b2c3d4e5f6...'
```

### GSTR-1 & GSTR-3B

```typescript
import { generateGstr1, validateGstr1, generateGstr3bFromGstr1 } from '@samavāya/wasm';

// Generate GSTR-1
const invoices = [...]; // Array of sales invoices
const gstr1 = await generateGstr1(invoices, '012024');  // January 2024
// → {
//     gstin: '27AAPFU0939F1ZV',
//     fp: '012024',
//     b2b: [...],
//     b2cl: [...],
//     b2cs: [...],
//     hsn: { data: [...] }
//   }

// Validate GSTR-1
const validation = await validateGstr1(gstr1);
// → {
//     valid: true,
//     errors: [],
//     summary: { totalInvoices: 150, totalValue: '5000000', totalTax: '450000' }
//   }

// Generate GSTR-3B from GSTR-1
const gstr3b = await generateGstr3bFromGstr1(gstr1);
// → {
//     gstin: '27AAPFU0939F1ZV',
//     ret_period: '012024',
//     sup_details: {
//       osup_det: { txval: '5000000', iamt: '200000', camt: '125000', samt: '125000' },
//       ...
//     },
//     itc_elg: {...}
//   }
```

---

## Crypto

### Hashing & HMAC

```typescript
import { sha256, sha512, hmacSha256, hmacSha512 } from '@samavāya/wasm';

// SHA-256
const hash = await sha256('Hello, World!');
// → 'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f'

// SHA-512
const hash512 = await sha512('Hello, World!');
// → '374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6c...'

// HMAC-SHA256
const hmac = await hmacSha256('secret-key', 'message');
// → 'a84b7c55a35a8c1c8e84c83f56c4f38cc95d7b2e...'
```

### Password Hashing

```typescript
import { hashPassword, verifyPassword } from '@samavāya/wasm';

// Hash password with PBKDF2
const hashed = await hashPassword('MySecurePassword123!', 100000);
// → {
//     hash: 'a1b2c3d4...',
//     salt: 'e5f6g7h8...',
//     iterations: 100000
//   }

// Verify password
const isValid = await verifyPassword(
  'MySecurePassword123!',
  hashed.hash,
  hashed.salt,
  hashed.iterations
);
// → true
```

### Encryption

```typescript
import { encryptAesGcm, decryptAesGcm, generateKey, generateApiKey } from '@samavāya/wasm';

// Generate 256-bit key
const key = await generateKey(32);
// → 'a1b2c3d4e5f6g7h8...' (64 hex chars)

// Encrypt
const encrypted = await encryptAesGcm('Sensitive data', key);
// → { ciphertext: '...', nonce: '...', tag: '...' }

// Decrypt
const decrypted = await decryptAesGcm(
  encrypted.ciphertext,
  key,
  encrypted.nonce,
  encrypted.tag
);
// → { plaintext: 'Sensitive data', success: true }

// Generate API key
const apiKey = await generateApiKey('sk');
// → 'sk_a1b2c3d4e5f6g7h8i9j0...'
```

---

## i18n (Localization)

### Number Formatting

```typescript
import {
  formatIndianNumber,
  formatCompactIndian,
  formatCurrency,
  parseIndianNumber
} from '@samavāya/wasm';

// Indian number system (lakhs, crores)
const formatted = await formatIndianNumber('12345678', 2);
// → '1,23,45,678.00'

// Compact notation
const compact = await formatCompactIndian('12345678');
// → '1.23 Cr'

const compact2 = await formatCompactIndian('1234567');
// → '12.35 L'

// Currency formatting
const currency = await formatCurrency('123456.50', 'INR', true);
// → '₹1,23,456.50'

const usd = await formatCurrency('123456.50', 'USD', true);
// → '$123,456.50'

// Parse formatted number
const parsed = await parseIndianNumber('1,23,45,678.00');
// → '12345678.00'
```

### Amount to Words

```typescript
import { amountToWords, amountToWordsHindi, formatCurrencyWords } from '@samavāya/wasm';

// English
const words = await amountToWords('123456.50');
// → 'One Lakh Twenty Three Thousand Four Hundred Fifty Six and Fifty Paise'

// With currency
const currencyWords = await formatCurrencyWords('123456.50', 'INR');
// → 'Rupees One Lakh Twenty Three Thousand Four Hundred Fifty Six and Fifty Paise Only'

// Hindi
const hindiWords = await amountToWordsHindi('123456');
// → 'एक लाख तेईस हज़ार चार सौ छप्पन'
```

### Date & Financial Year

```typescript
import { formatDate, formatRelativeTime, getFinancialYear, getQuarter } from '@samavāya/wasm';

// Format date
const date = await formatDate('2024-01-15', 'DD/MM/YYYY');
// → '15/01/2024'

const longDate = await formatDate('2024-01-15', 'D MMMM YYYY');
// → '15 January 2024'

// Relative time
const relative = await formatRelativeTime('2024-01-10', '2024-01-15');
// → '5 days ago'

// Financial year
const fy = await getFinancialYear('2024-01-15');
// → {
//     year: 2023,
//     label: 'FY 2023-24',
//     startDate: '2023-04-01',
//     endDate: '2024-03-31'
//   }

// Quarter
const quarter = await getQuarter('2024-01-15');
// → {
//     quarter: 4,
//     label: 'Q4 FY 2023-24',
//     startDate: '2024-01-01',
//     endDate: '2024-03-31'
//   }
```

---

## Offline Sync

### Change Tracking

```typescript
import { createChangeRecord, calculateDelta, applyDelta } from '@samavāya/wasm';

// Create change record
const change = await createChangeRecord(
  'invoice',           // entity type
  'INV-001',          // entity ID
  'update',           // operation
  { amount: '11000' }, // data
  'user-123'          // user ID
);
// → {
//     id: 'chg_a1b2c3...',
//     entityType: 'invoice',
//     entityId: 'INV-001',
//     operation: 'update',
//     timestamp: '2024-01-15T10:30:00Z',
//     syncStatus: 'pending',
//     checksum: 'abc123...'
//   }

// Calculate delta
const before = { name: 'ABC', amount: '10000', status: 'draft' };
const after = { name: 'ABC', amount: '11000', status: 'pending' };
const delta = await calculateDelta(before, after);
// → {
//     added: {},
//     modified: {
//       amount: { before: '10000', after: '11000' },
//       status: { before: 'draft', after: 'pending' }
//     },
//     removed: []
//   }

// Apply delta
const newState = await applyDelta(before, delta);
// → { name: 'ABC', amount: '11000', status: 'pending' }
```

### Conflict Resolution

```typescript
import { detectConflict, resolveConflict, threeWayMerge } from '@samavāya/wasm';

// Detect conflict
const conflict = await detectConflict(localChange, remoteChange);
// → {
//     id: 'conf_x1y2z3...',
//     conflictType: 'update_update',
//     conflictingFields: ['amount', 'status']
//   }

// Resolve with strategy
const resolved = await resolveConflict(localChange, remoteChange, 'server_wins');
// → { ...resolvedChangeRecord }

// Three-way merge
const base = { name: 'ABC', amount: '10000', status: 'draft' };
const local = { name: 'ABC Corp', amount: '10000', status: 'draft' };
const remote = { name: 'ABC', amount: '11000', status: 'pending' };

const merged = await threeWayMerge(base, local, remote);
// → {
//     merged: { name: 'ABC Corp', amount: '11000', status: 'pending' },
//     conflicts: [],
//     hasConflicts: false
//   }
```

### Version Vectors

```typescript
import {
  versionVectorIncrement,
  versionVectorMerge,
  versionVectorCompare
} from '@samavāya/wasm';

// Increment version
const v1 = { versions: { 'node-1': 1, 'node-2': 2 } };
const v2 = await versionVectorIncrement(v1, 'node-1');
// → { versions: { 'node-1': 2, 'node-2': 2 } }

// Merge vectors
const merged = await versionVectorMerge(v1, v2);
// → { versions: { 'node-1': 2, 'node-2': 2 } }

// Compare vectors
const comparison = await versionVectorCompare(v1, v2);
// → 'before' | 'after' | 'equal' | 'concurrent'
```

---

## Svelte Integration

### Basic Usage

```svelte
<script lang="ts">
  import { calculateGst, validateGstin } from '@samavāya/wasm';
  import type { GstResult } from '@samavāya/wasm';

  let amount = '10000';
  let rate = '18';
  let result: GstResult | null = null;

  async function calculate() {
    result = await calculateGst({
      amount,
      rate,
      isInclusive: false,
      sourceState: 'MH',
      destState: 'KA'
    });
  }
</script>

<input bind:value={amount} type="text" placeholder="Amount" />
<input bind:value={rate} type="text" placeholder="Rate" />
<button onclick={calculate}>Calculate GST</button>

{#if result}
  <div class="result">
    <p>IGST: ₹{result.igst}</p>
    <p>Total: ₹{result.totalAmount}</p>
  </div>
{/if}
```

### Reactive Validation with $effect

```svelte
<script lang="ts">
  import { validateGstin } from '@samavāya/wasm';
  import type { GstinValidationResult } from '@samavāya/wasm';

  let gstin = $state('');
  let validationResult: GstinValidationResult | null = $state(null);
  let isValidating = $state(false);

  $effect(() => {
    if (gstin.length === 15) {
      isValidating = true;
      validateGstin(gstin)
        .then(r => validationResult = r)
        .finally(() => isValidating = false);
    } else {
      validationResult = null;
    }
  });
</script>

<div class="form-field">
  <label for="gstin">GSTIN</label>
  <input
    id="gstin"
    bind:value={gstin}
    placeholder="27AAPFU0939F1ZV"
    maxlength="15"
  />

  {#if isValidating}
    <span class="validating">Validating...</span>
  {:else if validationResult}
    {#if validationResult.valid}
      <span class="valid">✓ {validationResult.stateName} - {validationResult.entityType}</span>
    {:else}
      <span class="invalid">✗ {validationResult.error}</span>
    {/if}
  {/if}
</div>

<style>
  .valid { color: green; }
  .invalid { color: red; }
  .validating { color: gray; }
</style>
```

### Form with Multiple WASM Calculations

```svelte
<script lang="ts">
  import {
    calculateGst,
    calculatePrice,
    formatIndianNumber,
    amountToWords
  } from '@samavāya/wasm';

  let basePrice = $state('1000');
  let quantity = $state('10');
  let discountPercent = $state('5');
  let gstRate = $state('18');
  let sourceState = $state('MH');
  let destState = $state('MH');

  let calculation = $state<{
    lineTotal: string;
    discountAmount: string;
    taxableAmount: string;
    gst: { cgst: string; sgst: string; igst: string };
    total: string;
    totalInWords: string;
  } | null>(null);

  async function calculate() {
    // Calculate price with discount
    const priceResult = await calculatePrice({
      basePrice,
      quantity,
      discounts: [{ discountType: 'percentage', value: discountPercent }],
      taxRate: '0',
      includeTax: false
    });

    // Calculate GST
    const gstResult = await calculateGst({
      amount: priceResult.netAmount,
      rate: gstRate,
      isInclusive: false,
      sourceState,
      destState
    });

    // Format total and convert to words
    const formattedTotal = await formatIndianNumber(gstResult.totalAmount, 2);
    const words = await amountToWords(gstResult.totalAmount);

    calculation = {
      lineTotal: priceResult.grossAmount,
      discountAmount: priceResult.discountAmount,
      taxableAmount: priceResult.netAmount,
      gst: {
        cgst: gstResult.cgst,
        sgst: gstResult.sgst,
        igst: gstResult.igst
      },
      total: formattedTotal,
      totalInWords: words
    };
  }
</script>

<form onsubmit|preventDefault={calculate}>
  <div class="grid">
    <input bind:value={basePrice} type="text" placeholder="Base Price" />
    <input bind:value={quantity} type="text" placeholder="Quantity" />
    <input bind:value={discountPercent} type="text" placeholder="Discount %" />
    <input bind:value={gstRate} type="text" placeholder="GST Rate" />
    <select bind:value={sourceState}>
      <option value="MH">Maharashtra</option>
      <option value="KA">Karnataka</option>
      <option value="DL">Delhi</option>
    </select>
    <select bind:value={destState}>
      <option value="MH">Maharashtra</option>
      <option value="KA">Karnataka</option>
      <option value="DL">Delhi</option>
    </select>
  </div>
  <button type="submit">Calculate</button>
</form>

{#if calculation}
  <table>
    <tr><td>Line Total</td><td>₹{calculation.lineTotal}</td></tr>
    <tr><td>Discount</td><td>₹{calculation.discountAmount}</td></tr>
    <tr><td>Taxable</td><td>₹{calculation.taxableAmount}</td></tr>
    {#if calculation.gst.igst !== '0'}
      <tr><td>IGST</td><td>₹{calculation.gst.igst}</td></tr>
    {:else}
      <tr><td>CGST</td><td>₹{calculation.gst.cgst}</td></tr>
      <tr><td>SGST</td><td>₹{calculation.gst.sgst}</td></tr>
    {/if}
    <tr class="total"><td>Total</td><td>₹{calculation.total}</td></tr>
  </table>
  <p class="words">{calculation.totalInWords}</p>
{/if}
```

### Payroll Calculator Component

```svelte
<script lang="ts">
  import {
    calculateIncomeTax,
    compareTaxRegimes,
    calculateAllStatutory,
    calculateCtcBreakdown
  } from '@samavāya/wasm';
  import type { TaxRegime, IncomeTaxResult, CtcBreakdown } from '@samavāya/wasm';

  let annualCtc = $state('1200000');
  let selectedState = $state('MH');
  let regime: TaxRegime = $state('new');

  let ctcBreakdown: CtcBreakdown | null = $state(null);
  let taxResult: IncomeTaxResult | null = $state(null);
  let regimeComparison: { oldRegime: IncomeTaxResult; newRegime: IncomeTaxResult; recommendation: TaxRegime; savings: string } | null = $state(null);

  async function calculate() {
    // Get CTC breakdown
    ctcBreakdown = await calculateCtcBreakdown(annualCtc, {
      includeGratuity: true,
      pfOnFullBasic: false
    });

    // Calculate income tax
    taxResult = await calculateIncomeTax({
      grossIncome: ctcBreakdown.annualGross,
      regime,
      age: 35,
      deductions: {
        section80C: '150000',
        standardDeduction: '50000'
      }
    });

    // Compare regimes
    regimeComparison = await compareTaxRegimes(ctcBreakdown.annualGross, {
      section80C: '150000',
      standardDeduction: '50000'
    });
  }
</script>

<div class="payroll-calculator">
  <h2>Payroll Calculator</h2>

  <div class="input-group">
    <label>Annual CTC</label>
    <input bind:value={annualCtc} type="text" />
  </div>

  <div class="input-group">
    <label>State</label>
    <select bind:value={selectedState}>
      <option value="MH">Maharashtra</option>
      <option value="KA">Karnataka</option>
      <option value="TN">Tamil Nadu</option>
      <option value="DL">Delhi</option>
    </select>
  </div>

  <div class="input-group">
    <label>Tax Regime</label>
    <div class="radio-group">
      <label>
        <input type="radio" bind:group={regime} value="old" /> Old Regime
      </label>
      <label>
        <input type="radio" bind:group={regime} value="new" /> New Regime
      </label>
    </div>
  </div>

  <button onclick={calculate}>Calculate</button>

  {#if ctcBreakdown && taxResult}
    <div class="results">
      <h3>CTC Breakdown (Monthly)</h3>
      <table>
        <tr><td>Basic</td><td>₹{(Number(ctcBreakdown.basic) / 12).toFixed(0)}</td></tr>
        <tr><td>HRA</td><td>₹{(Number(ctcBreakdown.hra) / 12).toFixed(0)}</td></tr>
        <tr><td>Special Allowance</td><td>₹{(Number(ctcBreakdown.specialAllowance) / 12).toFixed(0)}</td></tr>
        <tr><td>Employer PF</td><td>₹{(Number(ctcBreakdown.employerPf) / 12).toFixed(0)}</td></tr>
        <tr class="total"><td>Gross</td><td>₹{ctcBreakdown.monthlyGross}</td></tr>
      </table>

      <h3>Income Tax ({regime === 'new' ? 'New' : 'Old'} Regime)</h3>
      <table>
        <tr><td>Taxable Income</td><td>₹{taxResult.taxableIncome}</td></tr>
        <tr><td>Tax Before Cess</td><td>₹{taxResult.taxAfterRebate}</td></tr>
        <tr><td>Education Cess (4%)</td><td>₹{taxResult.educationCess}</td></tr>
        <tr class="total"><td>Total Tax</td><td>₹{taxResult.totalTax}</td></tr>
        <tr><td>Monthly TDS</td><td>₹{(Number(taxResult.totalTax) / 12).toFixed(0)}</td></tr>
        <tr><td>Effective Rate</td><td>{taxResult.effectiveRate}%</td></tr>
      </table>

      {#if regimeComparison}
        <div class="recommendation">
          <h4>Regime Comparison</h4>
          <p>Old Regime Tax: ₹{regimeComparison.oldRegime.totalTax}</p>
          <p>New Regime Tax: ₹{regimeComparison.newRegime.totalTax}</p>
          <p class="highlight">
            Recommended: <strong>{regimeComparison.recommendation === 'new' ? 'New' : 'Old'} Regime</strong>
            (Save ₹{regimeComparison.savings})
          </p>
        </div>
      {/if}
    </div>
  {/if}
</div>
```

### BOM Explosion Component

```svelte
<script lang="ts">
  import { explodeBom, calculateCostRollup } from '@samavāya/wasm';
  import type { BomItem, BomExplosionResult, BomCostResult } from '@samavāya/wasm';

  let selectedProduct = $state('FG-001');
  let quantity = $state('100');
  let explosion: BomExplosionResult | null = $state(null);
  let costs: BomCostResult | null = $state(null);

  // Sample BOM data
  const bomData: Record<string, BomItem[]> = {
    'FG-001': [
      { componentId: 'SA-001', componentName: 'Sub-Assembly A', quantity: '2', uom: 'NOS' },
      { componentId: 'RM-001', componentName: 'Steel Sheet', quantity: '0.5', uom: 'KG' }
    ],
    'SA-001': [
      { componentId: 'RM-002', componentName: 'Aluminum Bar', quantity: '0.25', uom: 'KG' },
      { componentId: 'RM-003', componentName: 'Bolts M6', quantity: '4', uom: 'NOS' }
    ]
  };

  const itemCosts: Record<string, string> = {
    'RM-001': '80',
    'RM-002': '150',
    'RM-003': '5'
  };

  async function explode() {
    explosion = await explodeBom(selectedProduct, bomData, quantity);
    costs = await calculateCostRollup(selectedProduct, bomData, itemCosts);
  }
</script>

<div class="bom-explorer">
  <h2>BOM Explosion</h2>

  <div class="controls">
    <select bind:value={selectedProduct}>
      <option value="FG-001">FG-001 - Finished Product</option>
    </select>
    <input bind:value={quantity} type="text" placeholder="Quantity" />
    <button onclick={explode}>Explode BOM</button>
  </div>

  {#if explosion}
    <h3>Multi-Level Explosion</h3>
    <table>
      <thead>
        <tr>
          <th>Level</th>
          <th>Component</th>
          <th>Qty Required</th>
          <th>UOM</th>
        </tr>
      </thead>
      <tbody>
        {#each explosion.items as item}
          <tr class="level-{item.level}">
            <td>{item.level}</td>
            <td style="padding-left: {item.level * 20}px">{item.componentName}</td>
            <td>{item.requiredQuantity}</td>
            <td>{item.uom}</td>
          </tr>
        {/each}
      </tbody>
    </table>

    <h3>Consolidated Requirements</h3>
    <table>
      <thead>
        <tr><th>Component</th><th>Total Qty</th><th>UOM</th></tr>
      </thead>
      <tbody>
        {#each Object.entries(explosion.summary) as [id, data]}
          <tr>
            <td>{id}</td>
            <td>{data.totalQuantity}</td>
            <td>{data.uom}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}

  {#if costs}
    <h3>Cost Analysis</h3>
    <table>
      <tr><td>Material Cost</td><td>₹{costs.materialCost}</td></tr>
      <tr><td>Labor Cost</td><td>₹{costs.laborCost}</td></tr>
      <tr><td>Overhead</td><td>₹{costs.overheadCost}</td></tr>
      <tr class="total"><td>Total Cost</td><td>₹{costs.totalCost}</td></tr>
    </table>
  {/if}
</div>
```

### Preloading WASM Modules

```typescript
// src/hooks.client.ts
import { initializeWasm, initializeAllWasm } from '@samavāya/wasm';

// Preload critical modules on app start
initializeWasm().catch(console.error);

// Or preload all modules if needed
// initializeAllWasm().catch(console.error);
```

### Using in SvelteKit Load Functions

```typescript
// src/routes/invoice/+page.ts
import type { PageLoad } from './$types';
import { calculateGst, validateGstin } from '@samavāya/wasm';

export const load: PageLoad = async ({ params }) => {
  // WASM functions work in load functions too
  const gstResult = await calculateGst({
    amount: '10000',
    rate: '18',
    isInclusive: false,
    sourceState: 'MH',
    destState: 'KA'
  });

  return {
    gst: gstResult
  };
};
```

### Custom WASM Hook

```svelte
<!-- lib/hooks/useWasm.svelte.ts -->
<script lang="ts" module>
  import { onMount } from 'svelte';

  export function useWasmFunction<T, R>(
    wasmFn: (input: T) => Promise<R>,
    initialInput?: T
  ) {
    let result = $state<R | null>(null);
    let error = $state<Error | null>(null);
    let loading = $state(false);

    async function execute(input: T): Promise<R | null> {
      loading = true;
      error = null;
      try {
        result = await wasmFn(input);
        return result;
      } catch (e) {
        error = e instanceof Error ? e : new Error(String(e));
        return null;
      } finally {
        loading = false;
      }
    }

    if (initialInput !== undefined) {
      onMount(() => execute(initialInput));
    }

    return {
      get result() { return result; },
      get error() { return error; },
      get loading() { return loading; },
      execute
    };
  }
</script>
```

Usage:

```svelte
<script lang="ts">
  import { useWasmFunction } from '$lib/hooks/useWasm.svelte';
  import { validateGstin } from '@samavāya/wasm';

  const { result, loading, error, execute } = useWasmFunction(validateGstin);

  let gstin = $state('');

  async function validate() {
    await execute(gstin);
  }
</script>

<input bind:value={gstin} />
<button onclick={validate} disabled={loading}>
  {loading ? 'Validating...' : 'Validate'}
</button>

{#if error}
  <p class="error">{error.message}</p>
{:else if result}
  <p class={result.valid ? 'valid' : 'invalid'}>
    {result.valid ? '✓ Valid' : '✗ Invalid'}
  </p>
{/if}
```

---

## Error Handling

```typescript
import { validateGstin, isWasmSupported } from '@samavāya/wasm';

// Check WASM support
if (!isWasmSupported()) {
  console.error('WebAssembly not supported');
}

// Handle validation errors
try {
  const result = await validateGstin(userInput);
  if (!result.valid) {
    showError(result.error);
  }
} catch (err) {
  // WASM module loading failed
  console.error('Validation failed:', err);
}
```

---

## Build Commands

```bash
# Production build
pnpm wasm:build

# Development build (faster, larger)
pnpm wasm:build:dev

# Build specific crate
cd packages/wasm
./build.ps1 -Crate tax-engine

# Clean all artifacts
pnpm wasm:clean

# Run Rust tests
cd packages/wasm
cargo test
```

---

## Performance Tips

1. **Preload critical modules** on app initialization
2. **Use bulk operations** when processing multiple items (e.g., `calculateGstBulk`)
3. **Cache module references** - modules are cached after first load
4. **Avoid round-trips** - batch multiple calculations when possible
5. **Use quick checks** (`isValidGstin`) instead of full validation when you only need boolean
6. **Run calculations in parallel** when operations are independent

---

## File Structure

```
packages/wasm/
├── Cargo.toml              # Rust workspace config
├── build.sh                # Unix build script
├── build.ps1               # Windows build script
├── package.json            # Node package config
├── tsconfig.json           # TypeScript config
├── crates/
│   ├── core/               # Shared utilities
│   ├── tax-engine/         # GST, TDS, TCS
│   ├── validation/         # ID validations
│   ├── barcode/            # QR, Code128, EAN
│   ├── ledger/             # Accounting logic
│   ├── pricing/            # Price calculations
│   ├── payroll/            # Income tax, PF, ESI, CTC
│   ├── bom/                # Bill of Materials, MRP
│   ├── depreciation/       # Asset depreciation
│   ├── compliance/         # e-Invoice, GSTR
│   ├── crypto/             # Hashing, encryption
│   ├── i18n/               # Localization
│   └── offline/            # Sync, conflicts
├── src/
│   ├── index.ts            # Main exports
│   ├── types.ts            # TypeScript types
│   ├── loader.ts           # WASM loader
│   ├── tax.ts              # Tax bindings
│   ├── validation.ts       # Validation bindings
│   ├── barcode.ts          # Barcode bindings
│   ├── ledger.ts           # Ledger bindings
│   ├── pricing.ts          # Pricing bindings
│   ├── core.ts             # Core bindings
│   ├── payroll.ts          # Payroll bindings
│   ├── bom.ts              # BOM bindings
│   ├── depreciation.ts     # Depreciation bindings
│   ├── compliance.ts       # Compliance bindings
│   ├── crypto.ts           # Crypto bindings
│   ├── i18n.ts             # i18n bindings
│   └── offline.ts          # Offline bindings
└── pkg/                    # Built WASM output
    ├── core/
    ├── tax-engine/
    ├── payroll/
    └── ...
```

---

## Adding New WASM Module

### Step 1: Create Crate Structure

```bash
cd packages/wasm/crates
mkdir my-module
```

```toml
# crates/my-module/Cargo.toml
[package]
name = "samavaya-my-module"
version.workspace = true
edition.workspace = true

[lib]
crate-type = ["cdylib", "rlib"]

[dependencies]
wasm-bindgen.workspace = true
serde.workspace = true
serde-wasm-bindgen.workspace = true
samavaya-core = { path = "../core" }
```

### Step 2: Implement Rust Code

```rust
// crates/my-module/src/lib.rs
use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};

#[wasm_bindgen(start)]
pub fn init() {
    console_error_panic_hook::set_once();
}

#[derive(Serialize, Deserialize)]
pub struct MyInput {
    pub value: String,
}

#[derive(Serialize, Deserialize)]
pub struct MyOutput {
    pub result: String,
}

#[wasm_bindgen]
pub fn my_function(input: JsValue) -> JsValue {
    let input: MyInput = serde_wasm_bindgen::from_value(input).unwrap();
    let output = MyOutput { result: input.value };
    serde_wasm_bindgen::to_value(&output).unwrap()
}
```

### Step 3: Add to Workspace

```toml
# packages/wasm/Cargo.toml
[workspace]
members = [
    # ... existing
    "crates/my-module",
]

[workspace.dependencies]
samavaya-my-module = { path = "crates/my-module" }
```

### Step 4: Create TypeScript Bindings

```typescript
// packages/wasm/src/my-module.ts
import { loadWasmModule, type WasmModuleName } from './loader';

interface MyModuleWasm {
  my_function: (input: unknown) => unknown;
}

let wasmModule: MyModuleWasm | null = null;

async function ensureLoaded(): Promise<MyModuleWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<MyModuleWasm>('my-module' as WasmModuleName);
  }
  return wasmModule;
}

export async function myFunction(value: string): Promise<string> {
  const wasm = await ensureLoaded();
  const result = wasm.my_function({ value }) as { result: string };
  return result.result;
}
```

### Step 5: Update Loader

```typescript
// packages/wasm/src/loader.ts
export type WasmModuleName =
  | 'core'
  // ... existing
  | 'my-module';

const MODULE_PATHS: Record<WasmModuleName, string> = {
  // ... existing
  'my-module': '../pkg/my-module/samavaya_my_module',
};
```

### Step 6: Export from Index

```typescript
// packages/wasm/src/index.ts
export * as myModule from './my-module';
export { myFunction } from './my-module';
```

### Step 7: Build & Test

```bash
pnpm wasm:build
cargo test -p samavaya-my-module
```

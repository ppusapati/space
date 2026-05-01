//! HSN (Harmonized System of Nomenclature) and SAC (Services Accounting Code) utilities
//!
//! Handles HSN/SAC code validation and GST rate lookup

use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// HSN/SAC code information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HsnInfo {
    /// HSN/SAC code
    pub code: String,
    /// Description
    pub description: String,
    /// GST rate
    pub gst_rate: String,
    /// Cess rate (if applicable)
    pub cess_rate: Option<String>,
    /// Whether it's a service code (SAC) or goods (HSN)
    pub is_service: bool,
    /// Chapter/Section
    pub chapter: String,
}

/// Common HSN codes with GST rates (sample data)
/// In production, this would be a database lookup
const COMMON_HSN_RATES: &[(&str, &str, &str, Option<&str>)] = &[
    // Chapter 1-5: Animals, Meat, Fish, Dairy
    ("0401", "Milk and cream", "0", None),
    ("0402", "Milk powder, condensed milk", "5", None),
    ("0406", "Cheese", "12", None),
    ("0407", "Eggs", "0", None),

    // Chapter 6-14: Vegetables, Fruits, Cereals
    ("0701", "Potatoes", "0", None),
    ("0702", "Tomatoes", "0", None),
    ("0713", "Dried leguminous vegetables", "0", None),
    ("1001", "Wheat", "0", None),
    ("1006", "Rice", "0", None),
    ("1101", "Wheat flour", "0", None),

    // Chapter 15-24: Edible oils, Preparations
    ("1507", "Soybean oil", "5", None),
    ("1509", "Olive oil", "5", None),
    ("1701", "Cane/beet sugar", "5", None),
    ("1704", "Sugar confectionery", "18", None),
    ("1806", "Chocolate", "18", None),
    ("1905", "Bread, pastry, cakes", "12", None),
    ("2106", "Food preparations", "18", None),
    ("2201", "Mineral water", "18", None),
    ("2202", "Beverages", "18", None),
    ("2401", "Tobacco leaves", "28", Some("0")),
    ("2402", "Cigars, cigarettes", "28", Some("5")),

    // Chapter 25-27: Minerals, Fuels
    ("2501", "Salt", "5", None),
    ("2701", "Coal", "5", Some("400/t")),
    ("2709", "Crude petroleum", "5", None),
    ("2710", "Petroleum products", "18", None),
    ("2711", "LPG", "5", None),

    // Chapter 28-38: Chemicals
    ("3004", "Medicaments", "12", None),
    ("3006", "Pharmaceutical goods", "12", None),
    ("3304", "Beauty products", "18", None),
    ("3305", "Hair preparations", "18", None),
    ("3401", "Soap", "18", None),
    ("3402", "Detergents", "18", None),

    // Chapter 39-40: Plastics, Rubber
    ("3923", "Plastic articles", "18", None),
    ("4011", "Rubber tyres", "28", None),

    // Chapter 41-43: Leather
    ("4202", "Leather bags", "18", None),
    ("4203", "Leather garments", "12", None),

    // Chapter 44-49: Wood, Paper
    ("4802", "Paper", "12", None),
    ("4901", "Printed books", "0", None),
    ("4902", "Newspapers", "0", None),

    // Chapter 50-63: Textiles
    ("5208", "Cotton fabrics", "5", None),
    ("5209", "Cotton fabrics (heavy)", "5", None),
    ("6109", "T-shirts", "5", None),
    ("6203", "Men's suits", "12", None),
    ("6204", "Women's suits", "12", None),

    // Chapter 64-67: Footwear
    ("6401", "Waterproof footwear", "18", None),
    ("6402", "Rubber/plastic footwear", "18", None),
    ("6403", "Leather footwear", "18", None),

    // Chapter 68-70: Stone, Ceramic, Glass
    ("6910", "Ceramic sanitary fixtures", "18", None),
    ("7013", "Glassware", "18", None),

    // Chapter 71: Jewellery
    ("7101", "Pearls", "3", None),
    ("7102", "Diamonds", "0.25", None),
    ("7106", "Silver", "3", None),
    ("7108", "Gold", "3", None),
    ("7113", "Jewellery", "3", None),

    // Chapter 72-83: Base Metals
    ("7208", "Steel sheets", "18", None),
    ("7304", "Steel tubes", "18", None),
    ("7318", "Screws, bolts", "18", None),
    ("7606", "Aluminium plates", "18", None),
    ("8302", "Base metal mountings", "18", None),

    // Chapter 84-85: Machinery, Electronics
    ("8414", "Air pumps, compressors", "18", None),
    ("8415", "Air conditioners", "28", None),
    ("8418", "Refrigerators", "18", None),
    ("8422", "Dishwashing machines", "18", None),
    ("8443", "Printers", "18", None),
    ("8471", "Computers", "18", None),
    ("8473", "Computer parts", "18", None),
    ("8504", "Transformers", "18", None),
    ("8507", "Batteries", "28", None),
    ("8517", "Telephones, smartphones", "18", None),
    ("8518", "Audio equipment", "18", None),
    ("8521", "Video players", "18", None),
    ("8523", "Recording media", "18", None),
    ("8528", "Monitors, TVs", "18", None),
    ("8536", "Electrical switches", "18", None),

    // Chapter 86-89: Vehicles
    ("8703", "Motor cars", "28", Some("1-22")),
    ("8704", "Commercial vehicles", "28", None),
    ("8711", "Motorcycles", "28", Some("3")),
    ("8712", "Bicycles", "12", None),

    // Chapter 90: Instruments
    ("9001", "Optical fibers", "18", None),
    ("9018", "Medical instruments", "12", None),

    // Chapter 91-92: Watches, Musical
    ("9101", "Wrist watches", "18", None),
    ("9102", "Watches (other)", "18", None),

    // Chapter 94-96: Furniture, Misc
    ("9401", "Seats", "18", None),
    ("9403", "Furniture", "18", None),
    ("9503", "Toys", "18", None),
    ("9504", "Video games", "28", None),
    ("9506", "Sports equipment", "18", None),
];

/// Common SAC codes for services
const COMMON_SAC_RATES: &[(&str, &str, &str)] = &[
    // Accounting, Auditing
    ("9971", "Financial services", "18"),
    ("998211", "Accounting services", "18"),
    ("998212", "Auditing services", "18"),
    ("998213", "Tax consultancy", "18"),

    // Legal
    ("998214", "Insolvency services", "18"),
    ("998215", "Legal services", "18"),

    // IT Services
    ("998311", "IT consulting", "18"),
    ("998312", "IT design and development", "18"),
    ("998313", "IT hosting and infrastructure", "18"),
    ("998314", "IT support services", "18"),

    // Research
    ("998331", "R&D services", "18"),

    // Advertising
    ("998361", "Advertising services", "18"),

    // Rental
    ("997211", "Rental of residential property", "0"),
    ("997212", "Rental of commercial property", "18"),
    ("997311", "Leasing machinery", "18"),

    // Construction
    ("9954", "Construction services", "18"),
    ("995411", "Residential construction", "12"),
    ("995421", "Commercial construction", "18"),

    // Telecommunications
    ("9984", "Telecom services", "18"),
    ("998411", "Fixed line services", "18"),
    ("998412", "Mobile services", "18"),

    // Transport
    ("9965", "Goods transport", "5"),
    ("996511", "Rail freight", "5"),
    ("996521", "Road freight", "5"),
    ("9964", "Passenger transport", "5"),

    // Hospitality
    ("9963", "Accommodation services", "12"),
    ("996311", "Hotel rooms (up to 1000)", "12"),
    ("996312", "Hotel rooms (1000-7500)", "18"),
    ("996313", "Hotel rooms (above 7500)", "28"),

    // Restaurant
    ("9963", "Food services", "5"),
    ("996331", "Restaurant (non-AC)", "5"),
    ("996332", "Restaurant (AC)", "18"),

    // Education
    ("9992", "Education services", "0"),
    ("999210", "Pre-school education", "0"),
    ("999220", "Primary education", "0"),
    ("999291", "Commercial training", "18"),

    // Healthcare
    ("9993", "Healthcare services", "0"),
    ("999311", "Hospital services", "0"),
    ("999312", "Medical services", "0"),

    // Entertainment
    ("9996", "Recreational services", "18"),
    ("999611", "Sports and recreation", "18"),

    // Professional
    ("9983", "Professional services", "18"),
    ("998311", "Management consulting", "18"),
    ("998312", "Business consulting", "18"),

    // Banking
    ("9971", "Banking services", "18"),
    ("997111", "Deposit services", "18"),
    ("997112", "Credit services", "18"),
    ("997113", "Payment services", "18"),

    // Insurance
    ("9971", "Insurance services", "18"),
    ("997131", "Life insurance", "18"),
    ("997132", "Non-life insurance", "18"),
];

/// Validate HSN code format
#[wasm_bindgen]
pub fn validate_hsn(code: &str) -> bool {
    let code = code.trim();

    // HSN codes are 2, 4, 6, or 8 digits
    if code.is_empty() {
        return false;
    }

    // Must be all digits
    if !code.chars().all(|c| c.is_ascii_digit()) {
        return false;
    }

    let len = code.len();
    len == 2 || len == 4 || len == 6 || len == 8
}

/// Validate SAC code format
#[wasm_bindgen]
pub fn validate_sac(code: &str) -> bool {
    let code = code.trim();

    // SAC codes are typically 4-6 digits
    if code.is_empty() {
        return false;
    }

    // Must be all digits
    if !code.chars().all(|c| c.is_ascii_digit()) {
        return false;
    }

    // SAC codes start with 99
    if !code.starts_with("99") {
        return false;
    }

    let len = code.len();
    len >= 4 && len <= 6
}

/// Get GST rate for HSN code
#[wasm_bindgen]
pub fn get_hsn_gst_rate(code: &str) -> String {
    let code = code.trim();

    // Try exact match first
    for (hsn, _, rate, _) in COMMON_HSN_RATES {
        if *hsn == code {
            return rate.to_string();
        }
    }

    // Try prefix match (4 digits, then 2 digits)
    let prefixes = [4, 2];
    for len in prefixes {
        if code.len() >= len {
            let prefix = &code[..len];
            for (hsn, _, rate, _) in COMMON_HSN_RATES {
                if *hsn == prefix {
                    return rate.to_string();
                }
            }
        }
    }

    // Default to 18% if not found
    "18".to_string()
}

/// Get GST rate for SAC code
#[wasm_bindgen]
pub fn get_sac_gst_rate(code: &str) -> String {
    let code = code.trim();

    // Try exact match
    for (sac, _, rate) in COMMON_SAC_RATES {
        if *sac == code {
            return rate.to_string();
        }
    }

    // Try prefix match
    for len in [6, 4] {
        if code.len() >= len {
            let prefix = &code[..len];
            for (sac, _, rate) in COMMON_SAC_RATES {
                if *sac == prefix {
                    return rate.to_string();
                }
            }
        }
    }

    // Default for services
    "18".to_string()
}

/// Look up HSN/SAC information
#[wasm_bindgen]
pub fn lookup_hsn_sac(code: &str) -> JsValue {
    let code = code.trim();

    // Check if it's a service code
    let is_service = code.starts_with("99");

    if is_service {
        for (sac, desc, rate) in COMMON_SAC_RATES {
            if *sac == code || code.starts_with(sac) {
                let info = HsnInfo {
                    code: code.to_string(),
                    description: desc.to_string(),
                    gst_rate: rate.to_string(),
                    cess_rate: None,
                    is_service: true,
                    chapter: sac[..2].to_string(),
                };
                return serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL);
            }
        }
    } else {
        for (hsn, desc, rate, cess) in COMMON_HSN_RATES {
            if *hsn == code || code.starts_with(hsn) {
                let info = HsnInfo {
                    code: code.to_string(),
                    description: desc.to_string(),
                    gst_rate: rate.to_string(),
                    cess_rate: cess.map(|c| c.to_string()),
                    is_service: false,
                    chapter: if code.len() >= 2 { code[..2].to_string() } else { "".to_string() },
                };
                return serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL);
            }
        }
    }

    // Return basic info if not found
    let info = HsnInfo {
        code: code.to_string(),
        description: "Unknown".to_string(),
        gst_rate: if is_service { "18".to_string() } else { "18".to_string() },
        cess_rate: None,
        is_service,
        chapter: if code.len() >= 2 { code[..2].to_string() } else { "".to_string() },
    };
    serde_wasm_bindgen::to_value(&info).unwrap_or(JsValue::NULL)
}

/// Determine if code is HSN (goods) or SAC (services)
#[wasm_bindgen]
pub fn is_service_code(code: &str) -> bool {
    code.trim().starts_with("99")
}

/// Get chapter from HSN code
#[wasm_bindgen]
pub fn get_hsn_chapter(code: &str) -> String {
    let code = code.trim();
    if code.len() >= 2 {
        code[..2].to_string()
    } else {
        code.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hsn_validation() {
        assert!(validate_hsn("8471"));
        assert!(validate_hsn("84713000"));
        assert!(!validate_hsn("847")); // Invalid length
        assert!(!validate_hsn("84a1")); // Non-numeric
    }

    #[test]
    fn test_sac_validation() {
        assert!(validate_sac("9971"));
        assert!(validate_sac("998211"));
        assert!(!validate_sac("1234")); // Doesn't start with 99
        assert!(!validate_sac("99")); // Too short
    }

    #[test]
    fn test_hsn_rate_lookup() {
        assert_eq!(get_hsn_gst_rate("8471"), "18"); // Computers
        assert_eq!(get_hsn_gst_rate("0401"), "0");  // Milk
        assert_eq!(get_hsn_gst_rate("8703"), "28"); // Cars
    }

    #[test]
    fn test_is_service() {
        assert!(is_service_code("9971"));
        assert!(is_service_code("998211"));
        assert!(!is_service_code("8471"));
    }
}

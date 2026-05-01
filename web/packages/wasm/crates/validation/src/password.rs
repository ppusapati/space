//! Password strength validation

use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

/// Password strength levels
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PasswordStrength {
    VeryWeak,
    Weak,
    Fair,
    Strong,
    VeryStrong,
}

impl PasswordStrength {
    pub fn score(&self) -> u8 {
        match self {
            PasswordStrength::VeryWeak => 1,
            PasswordStrength::Weak => 2,
            PasswordStrength::Fair => 3,
            PasswordStrength::Strong => 4,
            PasswordStrength::VeryStrong => 5,
        }
    }

    pub fn label(&self) -> &'static str {
        match self {
            PasswordStrength::VeryWeak => "Very Weak",
            PasswordStrength::Weak => "Weak",
            PasswordStrength::Fair => "Fair",
            PasswordStrength::Strong => "Strong",
            PasswordStrength::VeryStrong => "Very Strong",
        }
    }
}

/// Password validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PasswordResult {
    pub valid: bool,
    pub strength: String,
    pub score: u8,
    pub errors: Vec<String>,
    pub suggestions: Vec<String>,
    pub has_lowercase: bool,
    pub has_uppercase: bool,
    pub has_digit: bool,
    pub has_special: bool,
    pub length: usize,
}

/// Validate password strength
#[wasm_bindgen]
pub fn validate_password(password: &str, min_length: Option<u32>) -> JsValue {
    let result = validate_password_internal(password, min_length.unwrap_or(8) as usize);
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Internal password validation
pub fn validate_password_internal(password: &str, min_length: usize) -> PasswordResult {
    let mut errors = Vec::new();
    let mut suggestions = Vec::new();

    // Check characteristics
    let has_lowercase = password.chars().any(|c| c.is_ascii_lowercase());
    let has_uppercase = password.chars().any(|c| c.is_ascii_uppercase());
    let has_digit = password.chars().any(|c| c.is_ascii_digit());
    let has_special = password.chars().any(|c| !c.is_alphanumeric());
    let length = password.len();

    // Check minimum length
    if length < min_length {
        errors.push(format!("Password must be at least {} characters", min_length));
    }

    // Check for required character types
    if !has_lowercase {
        errors.push("Password must contain at least one lowercase letter".to_string());
        suggestions.push("Add a lowercase letter".to_string());
    }
    if !has_uppercase {
        errors.push("Password must contain at least one uppercase letter".to_string());
        suggestions.push("Add an uppercase letter".to_string());
    }
    if !has_digit {
        errors.push("Password must contain at least one digit".to_string());
        suggestions.push("Add a number".to_string());
    }
    if !has_special {
        suggestions.push("Add a special character for stronger security".to_string());
    }

    // Check for common patterns
    let password_lower = password.to_lowercase();
    let common_passwords = [
        "password", "123456", "qwerty", "abc123", "admin",
        "letmein", "welcome", "monkey", "dragon", "master",
    ];
    if common_passwords.iter().any(|p| password_lower.contains(p)) {
        errors.push("Password contains a common word or pattern".to_string());
        suggestions.push("Avoid common words and patterns".to_string());
    }

    // Check for sequential characters
    if has_sequential_chars(password) {
        suggestions.push("Avoid sequential characters like '123' or 'abc'".to_string());
    }

    // Check for repeated characters
    if has_repeated_chars(password) {
        suggestions.push("Avoid repeated characters like 'aaa'".to_string());
    }

    // Calculate strength score
    let mut score = 0u8;
    if length >= min_length { score += 1; }
    if length >= 12 { score += 1; }
    if has_lowercase { score += 1; }
    if has_uppercase { score += 1; }
    if has_digit { score += 1; }
    if has_special { score += 1; }
    if length >= 16 { score += 1; }

    let strength = match score {
        0..=2 => PasswordStrength::VeryWeak,
        3 => PasswordStrength::Weak,
        4 => PasswordStrength::Fair,
        5 => PasswordStrength::Strong,
        _ => PasswordStrength::VeryStrong,
    };

    PasswordResult {
        valid: errors.is_empty(),
        strength: strength.label().to_string(),
        score: strength.score(),
        errors,
        suggestions,
        has_lowercase,
        has_uppercase,
        has_digit,
        has_special,
        length,
    }
}

/// Check for sequential characters
fn has_sequential_chars(password: &str) -> bool {
    let sequences = [
        "0123456789",
        "9876543210",
        "abcdefghijklmnopqrstuvwxyz",
        "zyxwvutsrqponmlkjihgfedcba",
        "qwertyuiop",
        "asdfghjkl",
        "zxcvbnm",
    ];

    let password_lower = password.to_lowercase();
    sequences.iter().any(|seq| {
        for i in 0..seq.len().saturating_sub(2) {
            if password_lower.contains(&seq[i..i+3]) {
                return true;
            }
        }
        false
    })
}

/// Check for repeated characters
fn has_repeated_chars(password: &str) -> bool {
    let chars: Vec<char> = password.chars().collect();
    for i in 0..chars.len().saturating_sub(2) {
        if chars[i] == chars[i+1] && chars[i+1] == chars[i+2] {
            return true;
        }
    }
    false
}

/// Quick password strength check (returns score 1-5)
#[wasm_bindgen]
pub fn password_strength_score(password: &str) -> u8 {
    validate_password_internal(password, 8).score
}

/// Check if password meets minimum requirements
#[wasm_bindgen]
pub fn is_valid_password(password: &str) -> bool {
    validate_password_internal(password, 8).valid
}

/// Generate password strength indicator color
#[wasm_bindgen]
pub fn password_strength_color(password: &str) -> String {
    match password_strength_score(password) {
        1 => "#ff4d4d".to_string(), // Red
        2 => "#ff944d".to_string(), // Orange
        3 => "#ffd24d".to_string(), // Yellow
        4 => "#9fdf9f".to_string(), // Light green
        5 => "#4dff4d".to_string(), // Green
        _ => "#cccccc".to_string(), // Gray
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_weak_password() {
        let result = validate_password_internal("password", 8);
        assert!(!result.valid);
        assert_eq!(result.strength, "Very Weak");
    }

    #[test]
    fn test_strong_password() {
        let result = validate_password_internal("MyStr0ng!Pass#123", 8);
        assert!(result.valid);
        assert!(result.score >= 4);
    }

    #[test]
    fn test_missing_requirements() {
        let result = validate_password_internal("alllowercase", 8);
        assert!(!result.has_uppercase);
        assert!(!result.has_digit);
        assert!(!result.has_special);
    }

    #[test]
    fn test_sequential_chars() {
        assert!(has_sequential_chars("abc123"));
        assert!(!has_sequential_chars("aXc1Y3"));
    }

    #[test]
    fn test_repeated_chars() {
        assert!(has_repeated_chars("paassssword"));
        assert!(!has_repeated_chars("password"));
    }
}

export declare const formatCurrency: (amount: number, currency?: string, locale?: string, options?: Intl.NumberFormatOptions) => string;
export declare const formatPercent: (value: number, locale?: string, options?: Intl.NumberFormatOptions) => string;
export declare const formatNumber: (value: number, locale?: string, options?: Intl.NumberFormatOptions) => string;
export declare const formatBytes: (bytes: number, decimals?: number) => string;
export declare const formatDuration: (milliseconds: number) => string;
export declare const formatTime: (date: Date, format?: "12" | "24", includeSeconds?: boolean) => string;
export declare const formatPhoneNumber: (phoneNumber: string, format?: "US" | "INTERNATIONAL") => string;
export declare const formatCreditCard: (cardNumber: string, separator?: string) => string;
export declare const formatSSN: (ssn: string, masked?: boolean) => string;
export declare const formatAddress: (address: {
    street?: string;
    city?: string;
    state?: string;
    zipCode?: string;
    country?: string;
}, format?: "US" | "INTERNATIONAL") => string;
export declare const formatList: (items: string[], type?: "conjunction" | "disjunction", locale?: string) => string;
export declare const formatInitials: (name: string, maxInitials?: number) => string;
export declare const formatPlural: (count: number, singular: string, plural?: string) => string;
export declare const formatOrdinal: (number: number, locale?: string) => string;
export declare const formatFileSize: (bytes: number, binary?: boolean) => string;
export declare const formatHashtag: (text: string) => string;
export declare const formatMention: (username: string) => string;
export declare const formatTemplate: (template: string, variables: Record<string, any>) => string;
export declare const formatJson: (obj: any, indent?: number) => string;
export declare const formatCSV: (data: Record<string, any>[], delimiter?: string) => string;
export declare const formatMarkdown: {
    bold: (text: string) => string;
    italic: (text: string) => string;
    code: (text: string) => string;
    codeBlock: (text: string, language?: string) => string;
    link: (text: string, url: string) => string;
    image: (alt: string, src: string) => string;
    heading: (text: string, level: number) => string;
    quote: (text: string) => string;
    list: (items: string[], ordered?: boolean) => string;
};
export declare const stripFormatting: {
    html: (text: string) => string;
    markdown: (text: string) => string;
    whitespace: (text: string) => string;
    nonPrintable: (text: string) => string;
};
//# sourceMappingURL=index.d.ts.map
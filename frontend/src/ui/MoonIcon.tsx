export function MoonIcon({ size = 20 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <radialGradient id="moonGlow" cx="35%" cy="35%">
          <stop offset="0%" style={{ stopColor: '#ffffff', stopOpacity: 0.25 }} />
          <stop offset="100%" style={{ stopColor: '#ffffff', stopOpacity: 0 }} />
        </radialGradient>
      </defs>

      <circle cx="16" cy="16" r="15" fill="#b8b8b8" stroke="#808080" strokeWidth="0.5"/>

      <circle cx="9" cy="7" r="1.8" fill="#878787" opacity="0.8"/>
      <circle cx="13" cy="11" r="1.68" fill="#8f8f8f" opacity="0.75"/>

      <circle cx="24" cy="8" r="1.92" fill="#909090" opacity="0.76"/>

      <circle cx="6" cy="16" r="1.8" fill="#888888" opacity="0.77"/>
      <circle cx="8" cy="24" r="1.68" fill="#8a8a8a" opacity="0.72"/>

      <circle cx="16" cy="19" r="1.56" fill="#878787" opacity="0.74"/>

      <circle cx="26" cy="16" r="1.8" fill="#888888" opacity="0.75"/>
      <circle cx="25" cy="24" r="1.8" fill="#7f7f7f" opacity="0.78"/>

      <circle cx="11" cy="28" r="1.68" fill="#8a8a8a" opacity="0.73"/>

      <circle cx="21" cy="27" r="1.68" fill="#808080" opacity="0.7"/>
      <circle cx="18" cy="23" r="1.56" fill="#909090" opacity="0.72"/>

      <circle cx="14" cy="9" r="1.44" fill="#878787" opacity="0.68"/>
      <circle cx="22" cy="20" r="1.44" fill="#8a8a8a" opacity="0.67"/>

      <circle cx="16" cy="16" r="15" fill="url(#moonGlow)"/>
    </svg>
  );
}

"use client";

export function FireflyMark({ className = "w-6 h-6" }: { className?: string }) {
  return (
    <div className={`relative flex items-center justify-center ${className}`}>
      <span className="absolute inset-0 rounded-full bg-[#C9F27D]/40 blur-md animate-firefly-glow" />

      <svg
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className="relative z-10 w-full h-full text-[#C9F27D] drop-shadow-[0_0_8px_rgba(201,242,125,0.8)]"
      >
        <path
          d="M12 3L14.5 8.5L20 9L15.5 13L17 19L12 16L7 19L8.5 13L4 9L9.5 8.5L12 3Z"
          fill="currentColor"
          fillOpacity="0.25"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinejoin="round"
        />
        <circle cx="12" cy="12" r="3" fill="#C9F27D" />
      </svg>
    </div>
  );
}

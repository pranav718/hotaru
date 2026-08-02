import React from "react";

interface InteractiveLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  href: string;
  children: React.ReactNode;
  className?: string;
  icon?: React.ReactNode;
  showArrow?: boolean;
}

export function InteractiveLink({
  href,
  children,
  className = "",
  target = "_blank",
  rel = "noopener noreferrer",
  icon,
  showArrow = true,
  ...props
}: InteractiveLinkProps) {
  return (
    <a
      href={href}
      target={target}
      rel={rel}
      className={`group inline-flex items-center gap-1.5 text-zinc-400 hover:text-zinc-100 transition-colors duration-200 ${className}`}
      {...props}
    >
      {icon}
      <span className="relative inline-block after:content-[''] after:absolute after:left-0 after:bottom-[-2px] after:w-0 after:h-[1px] after:bg-current after:transition-[width] after:duration-300 after:ease-in-out group-hover:after:w-full">
        {children}
      </span>
      {showArrow && (
        <svg
          className="w-4 h-4 transition-transform duration-500 ease-in-out transform group-hover:rotate-[405deg]"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M7 17L17 7M17 7H7M17 7v10" />
        </svg>
      )}
    </a>
  );
}

export default InteractiveLink;

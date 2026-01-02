import * as React from "react";
import { cn } from "@/lib/utils";

function NavbarMain({ className, ...props }: React.ComponentProps<"nav">) {
  return (
    <nav
      className={cn(
        "bg-card text-card-foreground border-b border-foreground/10 sticky top-0 z-50",
        className,
      )}
      {...props}
    />
  );
}

function NavbarContainer({ className, ...props }: React.ComponentProps<"div">) {
  return <div className={cn("container mx-auto px-4", className)} {...props} />;
}

function NavbarContent({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      className={cn("flex items-center justify-between h-16", className)}
      {...props}
    />
  );
}

function NavbarBrand({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex items-center gap-8", className)} {...props} />
  );
}

function NavbarLogo({
  className,
  href = "/",
  children,
  ...props
}: React.ComponentProps<"a">) {
  return (
    <a
      href={href}
      className={cn(
        "text-lg font-semibold hover:text-foreground/80 transition-colors",
        className,
      )}
      {...props}
    >
      {children}
    </a>
  );
}

function NavbarActions({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex items-center gap-4", className)} {...props} />
  );
}

export {
  NavbarMain,
  NavbarContainer,
  NavbarContent,
  NavbarBrand,
  NavbarLogo,
  NavbarActions,
};

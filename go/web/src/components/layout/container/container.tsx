import * as React from "react";
import { cn } from "@/lib/utils";

function Container({ className, ...props }: React.ComponentProps<"div">) {
  return <div className={cn("container mx-auto px-4", className)} {...props} />;
}

export { Container };

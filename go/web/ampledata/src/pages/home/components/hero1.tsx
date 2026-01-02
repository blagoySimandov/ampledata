import { ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
interface Hero1Props {
  heading: string;
  description: string;
  buttons?: {
    primary?: {
      text: string;
      url: string;
    };
    secondary?: {
      text: string;
      url: string;
    };
  };
  image: {
    src: string;
    alt: string;
  };
  className?: string;
}
const Hero1 = ({
  heading = "Blocks Built With Shadcn & Tailwind",
  description = "Finely crafted components built with React, Tailwind and Shadcn UI. Developers can copy and paste these blocks directly into their project.",
  buttons = {
    primary: {
      text: "Discover all components",
      url: "https://www.shadcnblocks.com",
    },
    secondary: {
      text: "View on GitHub",
      url: "https://www.shadcnblocks.com",
    },
  },
  image = {
    src: "https://deifkwefumgah.cloudfront.net/shadcnblocks/block/placeholder-1.svg",
    alt: "Hero section demo image showing interface components",
  },
  className,
}: Hero1Props) => {
  return (
    <section className={cn("py-32", className)}>
      <div className="container">
        <div className="grid items-center gap-8 lg:grid-cols-2">
          <div className="flex flex-col items-center text-center lg:items-start lg:text-left">
            <h1 className="my-6 text-5xl font-bold text-pretty lg:text-7xl">
              {heading}
            </h1>
            <p className="mb-8 max-w-xl text-lg text-muted-foreground lg:text-2xl">
              {description}
            </p>
            <div className="flex w-full flex-col justify-center gap-3 sm:flex-row lg:justify-start">
              {buttons.primary && (
                <Button
                  className="w-full sm:w-auto px-8 py-3 text-lg"
                  render={<a href={buttons.primary.url} />}
                  nativeButton={false}
                >
                  {buttons.primary.text}
                </Button>
              )}
              {buttons.secondary && (
                <Button
                  variant="outline"
                  className="w-full sm:w-auto px-8 py-3 text-lg"
                  render={<a href={buttons.secondary.url} />}
                  nativeButton={false}
                >
                  {buttons.secondary.text}
                  <ArrowRight className="size-5" />
                </Button>
              )}
            </div>
          </div>
          <img
            src={image.src}
            alt={image.alt}
            className="max-h-100 w-full rounded-md object-cover"
          />
        </div>
      </div>
    </section>
  );
};
export { Hero1 };

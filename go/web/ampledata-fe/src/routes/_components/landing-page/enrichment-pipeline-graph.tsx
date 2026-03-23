import { ArrowRight } from "lucide-react";
import { useEffect, useState } from "react";
import { PIPELINE_ANIMATION_INTERVAL_MS, PIPELINE_NODES } from "./constants";

export function EnrichmentPipelineGraph() {
  const [activeIdx, setActiveIdx] = useState(0);

  useEffect(() => {
    const id = setInterval(
      () => setActiveIdx((prevIndex) => (prevIndex + 1) % PIPELINE_NODES.length),
      PIPELINE_ANIMATION_INTERVAL_MS,
    );
    return () => clearInterval(id);
  }, []);

  return (
    <div className="w-full max-w-[90rem] mx-auto flex flex-col lg:flex-row items-center justify-center gap-4 lg:gap-2">
      {PIPELINE_NODES.map((node, i) => {
        const Icon = node.icon;
        const isActive = i === activeIdx;
        const isPast = i < activeIdx;

        return [
          <div
            key={`${node.label}-node`}
            className="flex flex-col items-center w-full lg:w-[220px] xl:w-[240px] shrink-0"
          >
            <div
              className={`w-full max-w-[320px] lg:max-w-none rounded-3xl border-2 p-8 lg:p-6 bg-card transition-all duration-500 z-10 relative ${
                isActive
                  ? "border-primary shadow-2xl scale-105 ring-4 ring-primary/20 lg:-translate-y-2"
                  : isPast
                    ? "border-primary/50 shadow-md opacity-90"
                    : "border-border shadow-sm opacity-60 grayscale-[0.2]"
              }`}
            >
              <div
                className={`mx-auto w-20 h-20 lg:w-16 lg:h-16 rounded-2xl flex items-center justify-center transition-all duration-500 ${node.color} ${
                  isActive ? "shadow-lg scale-110" : ""
                }`}
              >
                <Icon
                  className={`size-10 lg:size-8 transition-transform duration-500 ${
                    isActive ? "scale-110" : ""
                  }`}
                />
              </div>
              <div
                className={`mt-5 lg:mt-4 text-center text-xl lg:text-lg font-black leading-tight transition-colors duration-300 ${
                  isActive ? "text-primary" : "text-foreground"
                }`}
              >
                {node.label}
              </div>
              <p className="mt-3 lg:mt-2 text-base lg:text-sm text-center text-muted-foreground leading-relaxed">
                {node.desc}
              </p>
            </div>
          </div>,
          i < PIPELINE_NODES.length - 1 && (
            <div
              key={`${node.label}-connector`}
              className="flex items-center justify-center py-2 lg:py-0 lg:w-6 xl:w-10 shrink-0 z-0"
            >
              <ArrowRight
                className={`size-10 lg:size-8 transition-all duration-500 rotate-90 lg:rotate-0 ${
                  isPast
                    ? "text-primary opacity-100"
                    : isActive
                      ? "text-primary/70 animate-pulse translate-y-2 lg:translate-y-0 lg:translate-x-2"
                      : "text-muted opacity-30"
                }`}
              />
            </div>
          ),
        ];
      })}
    </div>
  );
}

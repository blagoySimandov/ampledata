import { Card, CardContent } from "@/components/ui/card";
import { Database, Zap, TrendingUp } from "lucide-react";

export function HeroStats() {
  const stats = [
    {
      icon: Database,
      label: "Upload Rate",
      value: "10x faster",
      percentage: 85,
    },
    {
      icon: Zap,
      label: "Enrichment Speed",
      value: "98% faster",
      percentage: 98,
    },
    {
      icon: TrendingUp,
      label: "Data Accuracy",
      value: "99.9%",
      percentage: 100,
    },
  ];

  return (
    <div className="mt-16 lg:mt-24">
      <Card className="overflow-hidden">
        <div className="bg-gradient-to-br from-primary/20 via-primary/5 to-background p-8 lg:p-12">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            {stats.map((stat, index) => (
              <Card key={index}>
                <CardContent className="p-6">
                  <div className="flex items-center gap-3 mb-4">
                    <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
                      <stat.icon className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                      <div className="text-sm text-muted-foreground">
                        {stat.label}
                      </div>
                      <div className="text-2xl font-bold">{stat.value}</div>
                    </div>
                  </div>
                  <div className="h-2 bg-secondary rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full"
                      style={{ width: `${stat.percentage}%` }}
                    />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </Card>
    </div>
  );
}

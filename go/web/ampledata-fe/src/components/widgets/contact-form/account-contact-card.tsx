import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface Props {
  children: React.ReactNode;
}

export function AccountContactCard({ children }: Props) {
  return (
    <Card className="gap-0">
      <CardHeader className="border-b">
        <CardTitle className="text-sm font-semibold">Contact Support</CardTitle>
        <p className="text-sm text-muted-foreground">
          Have a question or need help? Send us a message and we'll get back
          to you shortly.
        </p>
      </CardHeader>
      <CardContent className="pt-4 space-y-4">{children}</CardContent>
    </Card>
  );
}

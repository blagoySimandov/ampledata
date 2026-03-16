import type { TierResponse, SubscriptionStatusResponse } from "@/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { formatDollars, formatOveragePrice } from "@/lib/formatters";
import { ArrowRight, Check } from "lucide-react";

interface Props {
  currentTier: TierResponse;
  newTier: TierResponse;
  subscription: SubscriptionStatusResponse;
  isPending: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

function proratedDiff(
  currentCents: number,
  newCents: number,
  periodEnd: string | null,
) {
  if (!periodEnd) return null;
  const daysRemaining = Math.max(
    0,
    (new Date(periodEnd).getTime() - Date.now()) / (1000 * 60 * 60 * 24),
  );
  const diff = ((newCents - currentCents) / 100) * (daysRemaining / 30);
  return diff > 0 ? diff.toFixed(2) : null;
}

export function UpgradeDialog({
  currentTier,
  newTier,
  subscription,
  isPending,
  onConfirm,
  onClose,
}: Props) {
  const prorated = proratedDiff(
    currentTier.monthly_price_cents,
    newTier.monthly_price_cents,
    subscription.current_period_end,
  );

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle className="text-xl font-black">
            Upgrade to {newTier.display_name}
          </DialogTitle>
          <DialogDescription>
            You'll be charged immediately for the prorated difference.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="flex items-center gap-3 text-sm">
            <span className="text-muted-foreground">{currentTier.display_name}</span>
            <ArrowRight className="size-4 shrink-0 text-muted-foreground" />
            <span className="font-semibold">{newTier.display_name}</span>
          </div>

          <ul className="space-y-2">
            <li className="flex items-center gap-2 text-sm">
              <Check className="size-4 text-primary shrink-0" />
              <span>
                <span className="font-semibold">
                  {newTier.included_tokens.toLocaleString()}
                </span>{" "}
                cells / month{" "}
                <span className="text-muted-foreground">
                  (was {currentTier.included_tokens.toLocaleString()})
                </span>
              </span>
            </li>
            <li className="flex items-center gap-2 text-sm">
              <Check className="size-4 text-primary shrink-0" />
              <span>
                <span className="font-semibold">
                  {formatOveragePrice(newTier.overage_price_cents_decimal)}
                </span>{" "}
                per extra cell{" "}
                <span className="text-muted-foreground">
                  (was {formatOveragePrice(currentTier.overage_price_cents_decimal)})
                </span>
              </span>
            </li>
          </ul>

          <div className="rounded-lg bg-muted px-4 py-3 space-y-1 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">New monthly price</span>
              <span className="font-semibold">
                {formatDollars(newTier.monthly_price_cents)} / month
              </span>
            </div>
            {prorated && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Charged today (prorated)</span>
                <span className="font-semibold">~${prorated}</span>
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={isPending}>
            {isPending ? "Upgrading..." : `Upgrade to ${newTier.display_name}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

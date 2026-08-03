import { Alert, type AlertProps } from "../atoms/Alert";
import { Button } from "../atoms/Button";

export type PwaNoticeProps = {
  offline?: boolean;
  updateAvailable?: boolean;
  offlineReady?: boolean;
  onUpdate?: () => void;
  onDismiss?: () => void;
  className?: string;
};

export function PwaNotice({
  offline,
  updateAvailable,
  offlineReady,
  onUpdate,
  onDismiss,
  className,
}: PwaNoticeProps) {
  if (!offline && !updateAvailable && !offlineReady) return null;

  const alertProps: Pick<AlertProps, "variant" | "role"> =
    updateAvailable || offline
      ? { variant: "destructive", role: "alert" }
      : { role: "status" };

  return (
    <Alert {...alertProps} className={className}>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p>
          {offline
            ? "You are offline. Live data and actions are unavailable."
            : null}
          {updateAvailable ? "A safer application update is ready." : null}
          {offlineReady
            ? "The application shell is ready to work offline."
            : null}
        </p>
        <div className="flex gap-2">
          {updateAvailable ? (
            <Button size="sm" onClick={onUpdate}>
              Update
            </Button>
          ) : null}
          {!offline ? (
            <Button size="sm" variant="ghost" onClick={onDismiss}>
              Dismiss
            </Button>
          ) : null}
        </div>
      </div>
    </Alert>
  );
}

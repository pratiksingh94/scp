import { cn } from "@/lib/utils";
import { Badge } from "./ui/badge";

type Actor = "client" | "server" | "both"

const actorStuff: Record<Actor, { label: string; badgeClass: string; borderClass: string; align: string }> = {
    client: {
        label: "CLIENT",
        badgeClass: "text-xs bg-client/10 text-client border-client/20",
        borderClass: "border-1-primary",
        align: "mr-auto"
    },

    server: {
        label: "SERVER",
        badgeClass: "text-xs bg-server/10 text-server border-server/20",
        borderClass: "border-1-server",
        align: "ml-auto"
    },

    both: {
        label: "BOTH",
        badgeClass: "text-xs bg-both/10 text-both border-both/20",
        borderClass: "border-1-warning",
        align: "mx-auto"
    }
}


export default function StepCard({ step }: { step: any }) {
    const config = actorStuff[step.actor as Actor] ?? actorStuff.both

    return (
        <div
        className={cn(
            "w-full md:w-[58%] rounded-lg border border-border border-1-2 bg-card p-4 flex flex-col gap-3",
            config.align,
            config.borderClass,
            "animate-in fade-in slide-in-from-bottom-2 duration-300"
        )}
        >
            <div className="flex items-center justify-between gap-2">
                <Badge variant="outline" className={config.badgeClass}>{config.label}</Badge>
                <span className="text-muted-foreground text-xs">step {step.step}</span>
            </div>


            <p className="text-sm font-medium text-foreground">{step.title}</p>

            {step.data && Object.keys(step.data).length > 0 && (
                <div className="bg-muted rounded-md p-3 flex flex-col gap-1">
                    {Object.entries(step.data).map(([k, v]) => (
                        <div key={k} className="flex gap-2 text-xs flex-wrap">
                            <span className="text-muted-foreground shrink-0">{k}:</span>
                            <span  className="text-foreground break-all">{String(v)}</span>
                        </div>
                    ))}
                </div>
            )}

            {step.annotation && (
                <p className="text-xs text-muted-foreground leading-relaxed">{step.annotation}</p>
            )}
        </div>
    )
}
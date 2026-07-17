export default function PhaseReader({ phase }: { phase: "handshake" | "data" }) {
    return (
        <div className="flex items-center gap-4 my-8">
            <div className="h-px flex-1 bg-border"/>
            <span className="text-xs font-medium text-muted-foreground tracking-widest uppercase">
                {phase === "handshake" ? "Handshake Phase" : "Data Phase" }
            </span>
            <div className="h-px flex-1 bg-border"/>
        </div>
    )
}
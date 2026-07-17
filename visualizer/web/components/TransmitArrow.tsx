import { cn } from "@/lib/utils"

export default function TransmitArrow({step}: {step: any}) {
    const isClientToServer = step.actor == "client"

    const packetName = step.data?.structure ? step.type.replace(/_/g, " ").replace("sent", "").trim() : step.title

    const packetType = step.data?.packet_type ?? ""

    return (
        <div className="my-4 px-2">
            <div className="flex items-center gap-3">

                <span  className={cn("text-xs w-16 text-right shrink-0 font-medium", isClientToServer ? "text-client" : "text-muted-foreground")}>CLIENT</span>
                <div className={cn("h-px flex-1", isClientToServer ? "bg-client/50" : "bg-border")}/>



                <div className="flex flex-col items-center gap-0.5 min-w-[160px] text-center shrink-0">
                    {/* fuck ass character  */}
                    {isClientToServer ? <span className="text-client text-xs">▼</span> : <span className="text-server text-xs">▼</span>}

                    <span className="text-xs font-medium text-foreground">
                        {step.data?.packet_type?.split(" ")[0] ?? ""}
                    </span>
                    <span className="text-xs text-muted-foreground">
                        {step.data?.packet_type ?? ""}
                    </span>

                    {step.data?.payload_size && (
                        <span className="text-xs text-muted-foreground">
                            {step.data.payload_size}
                        </span>
                    )}


                </div>


                <div className={cn("h-px flex-1", !isClientToServer ? "bg-server/50" : "bg-border")}/>

                <span className={cn("text-xs w-16 shrink-0 font-medium", !isClientToServer ? "text-server" : "text-muted-foreground")}>SERVER</span>
            </div>

            <div className="text-xs text-muted-foreground mt-1 text-center">
                {isClientToServer ? "client -> server" : "server -> client"}
            </div>
        </div>
    )
} 
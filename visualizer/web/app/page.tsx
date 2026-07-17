import SimulationView from "@/components/SimulationView"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

export default function Home() {
  return (
    <div className="flex flex-col gap-16">
      <div className="flex flex-col gap-4 pt-8">
        <p className="text-xs text-muted-foreground tracking-widest uppercase">
          SCP - Secure Channel Protocol
        </p>
        <h1 className="text-4xl font-bold tracking-tight">
          TLS, but {" "}
          <Tooltip>
            <TooltipTrigger render={<span className="text-primary underline cursor-pointer" />}>
            worse
            </TooltipTrigger>
            <TooltipContent>
              Okay worse because I'm making it for learning purposes and NOT as a replacement for TLS.
            </TooltipContent>
          </Tooltip>
          {/* <span className="text-primary">own.</span> */}
        </h1>
        <p className="text-muted-foreground max-w-xl leading-relaxed">
          SCP is a custom secure channel protocol implementing X25519 key exchange, PSK authentication, and ChaCha20-Poly1305 encryption made in Go.
        </p>
        <div className="flex gap-2 flex-wrap">
          {["X25519 ECDH", "ChaCha20-Poly1305", "HKDF-SHA256", "PSK Auth"].map(tag => (
            <span key={tag} className="text-xs border border-border rounded-full px-3 py-1 text-muted-foreground">
              {tag}
            </span>
          ))}
        </div>
      </div>


      <SimulationView/>
    </div>
  )
}
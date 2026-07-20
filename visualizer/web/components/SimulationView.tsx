"use client"

import { useEffect, useRef, useState } from "react"
import { Button } from "./ui/button"
import PhaseReader from "./PhaseReader"
import TransmitArrow from "./TransmitArrow"
import StepCard from "./StepCard"


const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/ws"

export default function SimulationView() {
    const [steps, setSteps] = useState<any[]>([])
    const [running, setRunning] = useState(false)
    const [done, setDone] = useState(false)
    const wsRef = useRef<WebSocket | null>(null)
    const bottomRef = useRef<HTMLDivElement | null>(null)

    useEffect(() => {
        if(steps.length > 0) {
            bottomRef.current?.scrollIntoView({ behavior: "smooth" })
        }
    }, [steps])

    function startSimulation() {
        if(wsRef.current) {
            wsRef.current.close()
        }

        setSteps([])
        setDone(false)
        setRunning(true)

        const ws = new WebSocket(WS_URL)
        wsRef.current = ws

        ws.onmessage = (e) => {
            const data = JSON.parse(e.data)
            if(data.type === "simulation_done") {
                setRunning(false)
                setDone(true)
                return
            }

            setSteps(prev => [...prev, data])
        }

        ws.onerror = () => {
            setRunning(false)
        }

        ws.onclose = () => {
            setRunning(false)
        }
    }


    let lastPhase = ""

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-end justify-between">
                <div className="flex flex-col gap-1">
                    <p className="text-muted-foreground text-xs uppercase tracking-widest">Demo</p>
                    <h2 className="text-2xl font-semibold">Handshake Simulation</h2>
                    <p className="text-sm text-muted-foreground max-w-lg">
                        Watch a totally real SCP session start with every cryptograpic operation, every packet, step by step.
                    </p>
                </div>
                <Button onClick={startSimulation} disabled={running} className="cursor-pointer shrink-0">
                    {running ? "Running..." : done ? "Run Again" : "Run Simulation"}
                </Button>
            </div>


            {steps.length === 0 && !running && (
                <div className="flex flex-col items-center justify-center py-24 text-center gap-3 border border-dashed border-border rounded-lg">
                    <p className="text-sm text-muted-foreground">
                        Press <span className="text-foreground">Run Simulation</span> to watch the handshake
                    </p>
                </div>
            )}


            <div className="flex flex-col gap-3">
                {steps.map((step, i) => {
                    const showPhaseHeader = step.phase !== lastPhase
                    if(showPhaseHeader) lastPhase = step.phase

                    return (
                        <div key={i}>
                            {showPhaseHeader && <PhaseReader phase={step.phase}/>}
                            {step.is_transmit ? <TransmitArrow step={step}/> : <StepCard step={step}/>}
                        </div>
                    )
                })}


                {done && (
                    <div className="flex flex-col items-cente py-8 gap-2 text-center">
                        <p className="text-sm text-success">Simulation Complete</p>
                        <p className="text-xs text-muted-foreground">All 26 steps executed, secure channel ready</p>
                    </div>
                )}


                <div ref={bottomRef}/>
            </div>
        </div>
    )
}
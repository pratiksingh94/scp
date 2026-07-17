
// import { GithubLogoIcon } from "@phosphor-icons/react";
import Link from "next/link";

export default function Navbar() {
    return (
        <nav className="border-b border-border">
            <div className="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
                <div className="flex items-center gap-8">
                    <span className="font-semibold text-sm text-foreground">
                        SCP<span className="text-primary">://</span>
                    </span>
                    <div className="flex items-center gap-6 text-sm text-muted-foreground">
                        <Link href="/" className="hover:text-foreground transition-colors">Simulation</Link>
                        <Link href="/spec" className="hover:text-foreground transition-colors">Spec</Link>
                    </div>
                </div>

                <a
                href="https://github.com/pratiksingh94/scp"
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-foreground transition-colors"
                >
                    GitHub
                </a>
            </div>
        </nav>
    )
}
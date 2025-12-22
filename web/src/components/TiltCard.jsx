import { motion, useMotionValue, useSpring, useTransform } from 'framer-motion';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
    return twMerge(clsx(inputs));
}

export default function TiltCard({ children, className, onClick }) {
    const x = useMotionValue(0);
    const y = useMotionValue(0);

    const mouseXSpring = useSpring(x);
    const mouseYSpring = useSpring(y);

    const rotateX = useTransform(mouseYSpring, [-0.5, 0.5], ["17.5deg", "-17.5deg"]);
    const rotateY = useTransform(mouseXSpring, [-0.5, 0.5], ["-17.5deg", "17.5deg"]);

    const handleMouseMove = (e) => {
        const rect = e.currentTarget.getBoundingClientRect();
        const width = rect.width;
        const height = rect.height;
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;
        const xPct = mouseX / width - 0.5;
        const yPct = mouseY / height - 0.5;
        x.set(xPct);
        y.set(yPct);
    };

    const handleMouseLeave = () => {
        x.set(0);
        y.set(0);
    };

    return (
        <motion.div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            style={{
                rotateY,
                rotateX,
                transformStyle: "preserve-3d",
            }}
            onClick={onClick}
            className={cn(
                "relative rounded-xl border border-white/10 bg-white/5 backdrop-blur-md shadow-2xl transition-all duration-200 ease-linear",
                "hover:shadow-violet-500/10 hover:border-violet-500/30",
                className
            )}
        >
            <div
                style={{
                    transform: "translateZ(50px)",
                }}
                className="w-full h-full"
            >
                {children}
            </div>

            {/* Gloss Effect */}
            <div
                className="absolute inset-0 rounded-xl"
                style={{
                    background: `radial-gradient(circle at 50% 0%, rgba(255,255,255,0.05) 0%, transparent 60%)`,
                    transform: "translateZ(1px)",
                    pointerEvents: 'none'
                }}
            />
        </motion.div>
    );
}

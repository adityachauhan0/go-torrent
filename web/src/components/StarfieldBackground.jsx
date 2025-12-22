import { useRef, useMemo } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { Points, PointMaterial } from '@react-three/drei';
import * as random from 'three/examples/jsm/math/MeshSurfaceSampler'; // Not used, will use custom generator

// Generate random points in a sphere
function generateStars(count = 5000, radius = 1.5) {
    const positions = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
        const r = radius * Math.cbrt(Math.random());
        const theta = Math.random() * 2 * Math.PI;
        const phi = Math.acos(2 * Math.random() - 1);
        const x = r * Math.sin(phi) * Math.cos(theta);
        const y = r * Math.sin(phi) * Math.sin(theta);
        const z = r * Math.cos(phi);
        positions[i * 3] = x;
        positions[i * 3 + 1] = y;
        positions[i * 3 + 2] = z;
    }
    return positions;
}

function Stars(props) {
    const ref = useRef();
    const spheres = useMemo(() => generateStars(6000, 2), []);

    useFrame((state, delta) => {
        // Basic rotation
        ref.current.rotation.x -= delta / 10;
        ref.current.rotation.y -= delta / 15;

        // Warp speed effect based on download speed (if passed in props)
        // For now subtle movement
    });

    return (
        <group rotation={[0, 0, Math.PI / 4]}>
            <Points ref={ref} positions={spheres} stride={3} frustumCulled={false} {...props}>
                <PointMaterial
                    transparent
                    color="#8b5cf6" // Violet accent
                    size={0.002}
                    sizeAttenuation={true}
                    depthWrite={false}
                />
            </Points>
        </group>
    );
}

export default function StarfieldBackground({ speed = 0 }) {
    return (
        <div className="fixed inset-0 -z-10 bg-background">
            <Canvas camera={{ position: [0, 0, 1] }}>
                <Stars speed={speed} />
            </Canvas>
        </div>
    );
}

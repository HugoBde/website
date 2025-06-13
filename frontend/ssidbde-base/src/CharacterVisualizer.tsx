import * as THREE from "three";
import { Canvas, useFrame, useLoader } from "@react-three/fiber";
import { GLTFLoader } from "three/examples/jsm/Addons.js";
import modelUrl from "./models/character.gltf?url";
import { useRef } from "react";

function Character() {

  const loader = useLoader(GLTFLoader, modelUrl);

  const modelRef = useRef<THREE.Mesh>(null);

  useFrame(() => {
    modelRef.current!.position.z = 3;
    modelRef.current!.position.y = -0.5;
    modelRef.current!.rotation.y += 0.02;
    modelRef.current!.rotation.x = 0.5;
  })

  return (
    <primitive ref={modelRef} object={loader.scene} />
  )
}

function CharacterVisualizer() {

  return (
    <Canvas className="border-black border-2 rounded-md max-h-[500px] max-w-[500px]">
      <ambientLight intensity={Math.PI / 2} />
      <spotLight position={[10, 10, 10]} angle={0.15} penumbra={1} decay={0} intensity={Math.PI} />
      <pointLight position={[-10, -10, -10]} decay={0} intensity={Math.PI} />
      <Character />
    </Canvas>
  );
}

export default CharacterVisualizer;

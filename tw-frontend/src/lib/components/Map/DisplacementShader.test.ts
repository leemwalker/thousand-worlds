import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock BabylonJS dependencies
const mockShaderMaterial = {
    setTexture: vi.fn(),
    setFloat: vi.fn(),
    backFaceCulling: false,
};

vi.mock("@babylonjs/core/Materials/shaderMaterial", () => ({
    ShaderMaterial: vi.fn(() => mockShaderMaterial),
}));

vi.mock("@babylonjs/core/Materials/Textures/texture", () => ({
    Texture: vi.fn(),
}));

import { DisplacementShader } from "./DisplacementShader";
import { Scene } from "@babylonjs/core/scene";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";

describe("DisplacementShader", () => {
    let scene: Scene;
    let heightmap: Texture;

    beforeEach(() => {
        scene = {} as Scene;
        heightmap = new Texture("url", scene);
        vi.clearAllMocks();
    });

    it("should create a ShaderMaterial with correct name and scene", () => {
        const shader = new DisplacementShader(scene);
        shader.createMaterial(heightmap);

        expect(ShaderMaterial).toHaveBeenCalledWith(
            "displacementShader",
            scene,
            expect.objectContaining({
                vertex: "displacement",
                fragment: "displacement",
            }),
            expect.any(Object)
        );
    });

    it("should set heightmap texture and scale uniforms", () => {
        const shader = new DisplacementShader(scene);
        const material = shader.createMaterial(heightmap, 2.5);

        expect(mockShaderMaterial.setTexture).toHaveBeenCalledWith("heightmap", heightmap);
        expect(mockShaderMaterial.setFloat).toHaveBeenCalledWith("scale", 2.5);
    });

    it("should allow updating heightmap texture", () => {
        const shader = new DisplacementShader(scene);
        const material = shader.createMaterial(heightmap);

        const newTexture = new Texture("new_url", scene);
        shader.updateHeightmap(newTexture);

        expect(mockShaderMaterial.setTexture).toHaveBeenCalledWith("heightmap", newTexture);
    });
});

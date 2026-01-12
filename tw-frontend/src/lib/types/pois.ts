
export enum POIType {
    MountainPeak = "mountain_peak",
    Volcano = "volcano",
    Canyon = "canyon",
    RiverMouth = "river_mouth",
    Lake = "lake",
    Valley = "valley",
    DeepOcean = "deep_ocean"
}

export interface PointOfInterest {
    id: string;
    type: POIType;
    name?: string;
    location: {
        x: number;
        y: number;
    };
    coordinates: {
        lat: number;
        lon: number;
    };
    elevation: number;
    importance: number;
    description?: string;
}

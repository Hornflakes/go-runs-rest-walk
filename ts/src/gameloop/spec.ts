import { Vector2D } from './geometry.js';

export const PlayerWidth = 100;
export const PlayerHeight = 100;
export const BulletWidth = 35;
export const BulletHeight = 3;

export const Player0SpawnX = 2500;
export const Player1SpawnX = -2500;
export const Player0FireRateMs = 180;
export const Player1FireRateMs = 300;

export const Player0Spawn: Vector2D = [Player0SpawnX, 0];
export const Player1Spawn: Vector2D = [Player1SpawnX, 0];
export const Player0Dir: Vector2D = [-1, 0];
export const Player1Dir: Vector2D = [1, 0];

export const BulletSpeedMs = 1.0;
export const TickTargetMicros = 16_000;
export const MicrosPerMs = 1000;

package com.hornflakes.gorunsrestwalk.gameloop;

public final class Rect {

    public double x;
    public double y;
    public double width;
    public double height;

    public Rect(double x, double y, double width, double height) {
        this.x = x;
        this.y = y;
        this.width = width;
        this.height = height;
    }

    public void setPosition(double x, double y) {
        this.x = x;
        this.y = y;
    }

    public boolean collides(Rect other) {
        if (this.x > other.x + other.width || other.x > this.x + this.width) {
            return false;
        }
        if (this.y > other.y + other.height || other.y > this.y + this.height) {
            return false;
        }
        return true;
    }
}

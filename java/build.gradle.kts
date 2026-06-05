plugins {
    java
    application
}

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
}

application {
    mainClass = "com.hornflakes.gorunsrestwalk.Main"
}

repositories {
    mavenCentral()
}

dependencies {
    implementation("io.netty:netty-all:4.2.15.Final")
}

tasks.named<JavaExec>("run") {
    standardOutput = System.out
    errorOutput = System.err
}



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
    mainClass.set("com.hornflakes.gorunsrestwalk.Main")
}

repositories {
    mavenCentral()
}

dependencies {
    implementation("org.eclipse.jetty:jetty-server:12.1.10")
    implementation("org.eclipse.jetty.websocket:jetty-websocket-jetty-server:12.1.10")
    implementation("org.eclipse.jetty.websocket:jetty-websocket-jetty-api:12.1.10")
    implementation("org.slf4j:slf4j-nop:2.0.17")
    implementation("com.fasterxml.jackson.core:jackson-databind:2.19.0")
}

tasks.withType<JavaCompile> {
    options.compilerArgs.addAll(listOf("--enable-preview"))
}

tasks.withType<JavaExec> {
    jvmArgs("--enable-preview")
}

tasks.withType<Test> {
    jvmArgs("--enable-preview")
}

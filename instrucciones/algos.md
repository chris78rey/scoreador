El desarrollo del simulador contempla la implementación de una serie de algoritmos matemáticos, probabilísticos y lógicos procesados en el núcleo de Golang. A continuación, se detalla qué algoritmos se pueden implementar y cómo funcionan:

**1. Algoritmos de Física y Cinemática del Balón**
*   **Qué implementan:** El movimiento aerodinámico realista del balón a través del aire, logrando que ruede, se eleve o curve su trayectoria de forma orgánica.
*   **Cómo funcionan:** Utilizan un resolvedor numérico de ecuaciones diferenciales vectoriales. La aceleración del balón se calcula en tiempo real restándole al vector de **gravedad** la fuerza del **arrastre aerodinámico** (fricción del aire) y sumándole la **fuerza Magnus** (la sustentación o curvatura provocada por el efecto de rotación de la pelota). Para esto, cruza variables como la densidad del aire, la masa y área transversal de la pelota, y el eje de rotación de la misma.

**2. Algoritmos de Fatiga Estructural por Bloques**
*   **Qué implementan:** El cansancio asimétrico del equipo, penalizando esquemas tácticos demasiado exigentes. A medida que un bloque se cansa (su valor se acerca a 1.0), el algoritmo degrada su precisión de pases, su velocidad y su intensidad defensiva.
*   **Cómo funcionan:** Mediante una función matemática que calcula el desgaste de un bloque de forma incremental en cada "tick" (segundo) de la simulación. El cálculo suma la fatiga acumulada anterior con dos factores principales: el **nivel de presión** configurado por el usuario (elevado a una potencia de 1.5, lo que penaliza exponencialmente la presión alta) y la **distancia espacial recorrida** por los jugadores de ese bloque en ese instante, multiplicados por coeficientes constantes calibrados en el motor.

**3. Algoritmos de Resolución Probabilística (Pases, Intercepciones y Tiros)**
*   **Qué implementan:** La lógica subyacente que decide si una acción de juego tiene éxito o falla.
*   **Cómo funcionan:** 
    *   **Pases interbloque:** Utilizan un sistema matricial de precisión (Matriz $\Phi$). El usuario define una tasa de precisión base para enviar el balón del bloque de origen "A" al bloque de destino "B", y este algoritmo altera matemáticamente los radios de éxito y la desviación física de la pelota.
    *   **Intercepciones:** Calculan la proximidad geométrica entre el balón y los bloques defensivos, evaluando también su capacidad de reacción para cortar trayectorias.
    *   **Tiros al arco:** Se resuelven bajo modelos probabilísticos de **Goles Esperados (xG)**, cruzando variables estadísticas como la calidad del pase previo (asistencia), la posición tridimensional del rematador en el campo y las capacidades de atajada del portero rival.

**4. Algoritmos de Analítica Táctica Avanzada (LBS y SGM)**
*   **Qué implementan:** La calificación de la calidad estratégica de cada pase completado para evaluar qué tan bien un equipo rompe las líneas rivales.
*   **Cómo funcionan:** 
    *   **Line Bypass Score (LBS):** Es una sumatoria matemática que cuenta cuántos defensores rivales fueron superados por un pase. El algoritmo verifica si la posición vertical ($y$) de cada defensor estaba entre la posición del pasador y la del receptor receptor; si es así, suma 1 punto.
    *   **Space Gain Metric (SGM):** Cuantifica la ganancia neta de espacio libre. El algoritmo compara numéricamente la densidad defensiva rival que asfixiaba al pasador en el origen frente al espacio libre con el que cuenta el receptor tras recibir el balón.
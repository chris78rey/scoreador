La implementación de la lógica gratuita y premium se estructura bajo un modelo de monetización freemium, el cual separa estrictamente el acceso a la capacidad de procesamiento masivo y a la profundidad analítica avanzada, sin sacrificar la experiencia visual en la versión gratuita.

Esta separación se implementaría funcionalmente de la siguiente manera:

**Versión Básica (Acceso Gratuito)**
Está orientada a atraer a los entusiastas del fútbol que desean experimentar y observar el desarrollo de tácticas. A nivel de sistema, sus límites son:
*   **Simulación individual y visual:** Solo permite ejecutar **un partido a la vez** y cuenta con un límite en la velocidad de aceleración de tiempo (hasta un máximo de 5x). Esta versión incluye la visualización gráfica completa de las jugadas en pantalla.
*   **Análisis simplificado:** Los datos que arroja el juego se reducen a estadísticas tradicionales (como goles, tiros y posesión) y gráficos estáticos simples que muestran la fatiga de los bloques.
*   **Límites de guardado y progresión:** Restringe el almacenamiento local a un máximo de **3 plantillas de formación**. Además, ofrece un acceso de solo lectura a las tácticas de la comunidad, insignias de experiencia básica y retos semanales estándar.

**Versión Premium (Acceso por Suscripción)**
Está diseñada para transformar el juego en un simulador de nivel profesional para usuarios exigentes y analistas tácticos. Implementa los siguientes desbloqueos lógicos:
*   **Simulación Masiva Instantánea (Laboratorio):** Desbloquea la capacidad del motor en Golang para procesar cientos o miles de partidos instantáneos y ejecutar torneos completos en segundo plano sin renderizado visual, permitiendo generar análisis predictivos.
*   **Métricas Matemáticas Avanzadas:** Habilita el cálculo y la visualización de algoritmos tácticos complejos, tales como los Goles Esperados (xG), el *Line Bypass Score* (LBS) y la *Space Gain Metric* (SGM) por cada pase, así como mapas de control de espacio y EPV.
*   **Ecosistema Comunitario y Almacenamiento:** Permite el **guardado ilimitado** de plantillas e historiales de alineaciones, y da acceso activo a la red comunitaria para exportar e intercambiar tácticas de forma ilimitada.
*   **Desafíos Exclusivos:** El usuario puede acceder a laboratorios experimentales más profundos y a escenarios históricos que simulan el enfrentamiento contra equipos de leyenda.
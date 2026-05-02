-- Migración 008: noticias iniciales para que la home no quede vacía.
-- Idempotente por título único (en este seed los títulos son distintos).

INSERT INTO news (title, summary, category, category_class, published_at, image_url) VALUES
('Fuerte tormenta deja más de 50cm en Pirineos',
 'Las estaciones aragonesas y catalanas se preparan para un fin de semana espectacular tras las intensas precipitaciones...',
 'Nevada', 'nevada', '2026-03-14',
 'https://images.unsplash.com/photo-1513342774453-5d76a9768b41?q=80&w=400&h=225&auto=format&fit=crop'),
('Cómo elegir tus botas de esquí ideales',
 'La comodidad es clave. Te damos los 5 aspectos fundamentales en los que fijarte antes de comprar o alquilar material...',
 'Consejos', 'consejos', '2026-03-12',
 'https://images.unsplash.com/photo-1551698618-1dfe5d97d256?q=80&w=400&h=225&auto=format&fit=crop'),
('Campeonato Nacional de Slalom en Sierra Nevada',
 'Este fin de semana se reúnen los mejores atletas nacionales en la mítica pista del Río para disputarse el título...',
 'Evento', 'evento', '2026-03-10',
 'https://plus.unsplash.com/premium_photo-1664302791901-52c6159eaf78?q=80&w=400&h=225&auto=format&fit=crop'),
('Nuevos forfaits conjuntos para la temporada',
 'Se anuncian los nuevos pases de temporada que permitirán esquiar en diferentes comunidades autónomas con descuentos...',
 'General', 'general', '2026-03-08',
 'https://images.unsplash.com/photo-1486684338211-1a7ced564b0d?q=80&w=400&h=225&auto=format&fit=crop'),
('Alerta por riesgo de aludes nivel 4 en Andorra',
 'Tras las recientes precipitaciones y el aumento de las temperaturas, Protección Civil advierte del alto riesgo fuera de pistas...',
 'Nevada', 'nevada', '2026-03-05',
 'https://images.unsplash.com/photo-1732692583018-2345548e4e5a?q=80&w=400&h=225&auto=format&fit=crop'),
('Preparación física pre-temporada: Evita lesiones',
 'Una buena rutina de fuerza y flexibilidad en piernas y core es esencial para disfrutar de la nieve de forma segura...',
 'Consejos', 'consejos', '2026-03-01',
 'https://images.unsplash.com/photo-1596473536056-91eadf31189e?q=80&w=400&h=225&auto=format&fit=crop')
;
-- (sin ON CONFLICT: si vuelves a sembrar duplicas, así que esta migración
-- queda registrada en schema_migrations y no se vuelve a ejecutar)
